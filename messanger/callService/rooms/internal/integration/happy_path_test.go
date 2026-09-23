package integration_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"rooms/internal/kafka"
	"rooms/internal/migrator"
	"rooms/internal/model"
	participantRepo "rooms/internal/repository/participant"
	roomRepo "rooms/internal/repository/room"
	"rooms/internal/repository/roomcache"
	"rooms/internal/service/authctx"
	participantService "rooms/internal/service/participant"
	roomService "rooms/internal/service/room"
)

type stubJoinToken struct{}

func (stubJoinToken) GenerateJoinToken(_ context.Context, info tokens.JoinInfo) (string, error) {
	return "join-" + info.GetUserUUID() + "-" + info.GetRoomUUID(), nil
}

func (stubJoinToken) VerifyJoinToken(context.Context, string) (*tokens.JoinClaims, error) {
	return &tokens.JoinClaims{}, nil
}

type stubTURNConfig struct {
	secret string
	urls   []string
	ttl    time.Duration
}

func (c stubTURNConfig) SharedSecret() string { return c.secret }
func (c stubTURNConfig) URLs() []string       { return c.urls }
func (c stubTURNConfig) TTL() time.Duration   { return c.ttl }

type HappyPathSuite struct {
	suite.Suite

	ctx       context.Context
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
}

func TestHappyPathSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if !dockerAvailable() {
		t.Skip("docker is not available, skipping integration test")
	}
	suite.Run(t, new(HappyPathSuite))
}

func dockerAvailable() (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	provider, err := testcontainers.ProviderType(0).GetProvider()
	if err != nil || provider == nil {
		return false
	}
	_ = provider.Close()
	return true
}

func (s *HappyPathSuite) SetupSuite() {
	logger.SetNopLogger()
	s.ctx = context.Background()

	container, err := postgres.Run(
		s.ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("rooms"),
		postgres.WithUsername("rooms"),
		postgres.WithPassword("rooms"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	s.Require().NoError(err)
	s.container = container

	uri, err := container.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)

	pool, err := pgxpool.New(s.ctx, uri)
	s.Require().NoError(err)
	s.Require().NoError(pool.Ping(s.ctx))
	s.pool = pool

	_, thisFile, _, ok := runtime.Caller(0)
	s.Require().True(ok)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "migrations")

	m := migrator.NewMigrator(stdlib.OpenDBFromPool(pool), migrationsDir)
	s.Require().NoError(m.Up())
}

func (s *HappyPathSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		_ = s.container.Terminate(s.ctx)
	}
}

func (s *HappyPathSuite) TestCreateJoinLeaveEnd() {
	token := stubJoinToken{}
	rooms := roomRepo.NewRoomRepository(s.pool, token)
	parts := participantRepo.NewParticipantRepository(s.pool, token)
	cache := roomcache.NewNoop()

	rs := roomService.NewService(rooms, kafka.NoopProducer{}, cache, time.Hour)
	ps := participantService.NewService(
		rooms,
		parts,
		kafka.NoopProducer{},
		cache,
		token,
		stubTURNConfig{
			secret: "turn-test-secret",
			urls:   []string{"turn:localhost:3478"},
			ttl:    time.Hour,
		},
	)

	ownerUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()
	ownerCtx := authctx.WithUserID(s.ctx, ownerUUID)
	userCtx := authctx.WithUserID(s.ctx, userUUID)

	createResp, err := rs.CreateRoom(ownerCtx, model.CreateRoomRequest{
		Name:      "integration-room",
		OwnerUUID: ownerUUID,
		RoomSettings: model.RoomSettings{
			MaxParticipants: 5,
			Quality:         "HD",
			AutoClose:       false,
		},
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(createResp.Room.RoomUUID)
	s.Require().NotEmpty(createResp.JoinToken)
	roomUUID := createResp.Room.RoomUUID

	getResp, err := rs.GetRoom(ownerCtx, model.GetRoomRequest{RoomUUID: roomUUID})
	s.Require().NoError(err)
	s.Require().Equal("integration-room", getResp.Room.Name)
	s.Require().Equal("active", getResp.Room.Status)

	addResp, err := ps.AddParticipant(userCtx, model.AddParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
		Metadata: model.ParticipantMetadata{
			DisplayName: "tester",
			Role:        "participant",
		},
	})
	s.Require().NoError(err)
	s.Require().True(addResp.Success)
	s.Require().NotEmpty(addResp.JoinToken)

	isResp, err := ps.IsParticipant(userCtx, model.IsParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.Require().NoError(err)
	s.Require().True(isResp.IsActive)

	listResp, err := ps.GetParticipant(ownerCtx, model.GetParticipantsRequest{
		RoomUUID:   roomUUID,
		OnlyActive: true,
	})
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(listResp.ActiveCount, int32(1))

	_, err = ps.RemoveParticipant(userCtx, model.RemoveParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.Require().NoError(err)

	_, err = rs.EndRoom(ownerCtx, model.EndRoomRequest{
		RoomUUID:  roomUUID,
		OwnerUUID: ownerUUID,
	})
	s.Require().NoError(err)

	ended, err := rs.GetRoom(ownerCtx, model.GetRoomRequest{RoomUUID: roomUUID})
	s.Require().NoError(err)
	s.Require().Equal("ended", ended.Room.Status)
}
