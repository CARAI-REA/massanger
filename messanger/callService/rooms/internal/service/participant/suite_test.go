package participant

import (
	"context"
	"testing"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	"github.com/stretchr/testify/suite"

	"rooms/internal/kafka"
	"rooms/internal/repository/mocks"
	"rooms/internal/repository/roomcache"
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

func (c stubTURNConfig) SharedSecret() string   { return c.secret }
func (c stubTURNConfig) URLs() []string         { return c.urls }
func (c stubTURNConfig) TTL() time.Duration     { return c.ttl }

type ServiceSuite struct {
	suite.Suite

	ctx context.Context

	roomRepository        *mocks.RoomRepository
	participantRepository *mocks.ParticipantRepository

	service *service
}

func (s *ServiceSuite) SetupTest() {
	s.ctx = context.Background()
	s.roomRepository = mocks.NewRoomRepository(s.T())
	s.participantRepository = mocks.NewParticipantRepository(s.T())

	s.service = NewService(
		s.roomRepository,
		s.participantRepository,
		kafka.NoopProducer{},
		roomcache.NewNoop(),
		stubJoinToken{},
		stubTURNConfig{
			secret: "turn-test-secret",
			urls:   []string{"turn:localhost:3478"},
			ttl:    time.Hour,
		},
	)
}

func (s *ServiceSuite) TearDownTest() {
}

func TestServiceIntegration(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}
