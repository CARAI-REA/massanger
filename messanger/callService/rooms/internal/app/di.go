package app

import (
	"context"
	"fmt"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/closer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	jwtTokens "github.com/CARAI-REA/messanger/callService/platform/pkg/tokens/jwt"
	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"

	roomV1API "rooms/internal/api/room/v1"
	"rooms/internal/config"
	"rooms/internal/kafka"
	"rooms/internal/migrator"
	"rooms/internal/rediswire"
	"rooms/internal/repository"
	participantRepository "rooms/internal/repository/participant"
	roomRepository "rooms/internal/repository/room"
	"rooms/internal/service"
	participantService "rooms/internal/service/participant"
	roomService "rooms/internal/service/room"
)

type diContainer struct {
	roomsV1API roomsV1.RoomServiceServer

	roomService        service.RoomService
	participantService service.ParticipantService

	roomRepository        repository.RoomRepository
	participantRepository repository.ParticipantRepository
	roomCache             repository.RoomCache

	postgresClient *pgxpool.Pool
	redisPool      *redigo.Pool

	eventProducer    kafka.EventProducer
	joinTokenService tokens.JoinTokenService
	accessVerifier   tokens.AccessTokenVerifier
	serviceVerifier  tokens.ServiceTokenVerifier
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) RoomsV1API(ctx context.Context) roomsV1.RoomServiceServer {
	if d.roomsV1API == nil {
		d.roomsV1API = roomV1API.NewAPI(d.RoomService(ctx), d.ParticipantService(ctx))
	}
	return d.roomsV1API
}

func (d *diContainer) RoomService(ctx context.Context) service.RoomService {
	if d.roomService == nil {
		d.roomService = roomService.NewService(
			d.RoomRepository(ctx),
			d.EventProducer(ctx),
			d.RoomCache(ctx),
			config.AppConfig().Redis.CacheTTL(),
		)
	}
	return d.roomService
}

func (d *diContainer) ParticipantService(ctx context.Context) service.ParticipantService {
	if d.participantService == nil {
		d.participantService = participantService.NewService(
			d.RoomRepository(ctx),
			d.ParticipantRepository(ctx),
			d.EventProducer(ctx),
			d.RoomCache(ctx),
			d.JoinTokenService(),
			config.AppConfig().TURN,
		)
	}
	return d.participantService
}

func (d *diContainer) RoomRepository(ctx context.Context) repository.RoomRepository {
	if d.roomRepository == nil {
		d.roomRepository = roomRepository.NewRoomRepository(d.PostgresClient(ctx), d.JoinTokenService())
	}
	return d.roomRepository
}

func (d *diContainer) ParticipantRepository(ctx context.Context) repository.ParticipantRepository {
	if d.participantRepository == nil {
		d.participantRepository = participantRepository.NewParticipantRepository(
			d.PostgresClient(ctx),
			d.JoinTokenService(),
		)
	}
	return d.participantRepository
}

func (d *diContainer) RoomCache(ctx context.Context) repository.RoomCache {
	if d.roomCache == nil {
		d.roomCache = rediswire.NewRoomCache(d.RedisPool(), rediswire.PoolConfig{
			Host:              config.AppConfig().Redis.Host(),
			Port:              config.AppConfig().Redis.Port(),
			Password:          config.AppConfig().Redis.Password(),
			ConnectionTimeout: config.AppConfig().Redis.ConnectionTimeout(),
			MaxIdle:           config.AppConfig().Redis.MaxIdle(),
			IdleTimeout:       config.AppConfig().Redis.IdleTimeout(),
		})
	}
	return d.roomCache
}

func (d *diContainer) PostgresClient(ctx context.Context) *pgxpool.Pool {
	if d.postgresClient == nil {
		pool, err := pgxpool.New(ctx, config.AppConfig().Postgres.URI())
		if err != nil {
			panic(fmt.Sprintf("failed to connect to PostgreSQL: %s\n", err.Error()))
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			panic(fmt.Sprintf("failed to ping PostgreSQL: %v\n", err))
		}

		migrationDir := config.AppConfig().Postgres.MigrationDir()
		if migrationDir != "" {
			migratorRunner := migrator.NewMigrator(stdlib.OpenDBFromPool(pool), migrationDir)
			if err := migratorRunner.Up(); err != nil {
				panic(fmt.Sprintf("failed to run migrations: %v\n", err))
			}
		}

		closer.AddNamed("PostgreSQL pool", func(ctx context.Context) error {
			d.postgresClient.Close()
			return nil
		})

		d.postgresClient = pool
	}
	return d.postgresClient
}

func (d *diContainer) RedisPool() *redigo.Pool {
	if d.redisPool == nil {
		pool := rediswire.NewPool(rediswire.PoolConfig{
			Host:              config.AppConfig().Redis.Host(),
			Port:              config.AppConfig().Redis.Port(),
			Password:          config.AppConfig().Redis.Password(),
			ConnectionTimeout: config.AppConfig().Redis.ConnectionTimeout(),
			MaxIdle:           config.AppConfig().Redis.MaxIdle(),
			IdleTimeout:       config.AppConfig().Redis.IdleTimeout(),
		})

		closer.AddNamed("Redis pool", func(ctx context.Context) error {
			return d.redisPool.Close()
		})

		d.redisPool = pool
	}
	return d.redisPool
}

func (d *diContainer) EventProducer(ctx context.Context) kafka.EventProducer {
	if d.eventProducer == nil {
		producer, err := kafka.NewEventProducer(
			config.AppConfig().Kafka.Brokers(),
			config.AppConfig().Kafka.RoomEventsTopic(),
			config.AppConfig().Kafka.ParticipantEventsTopic(),
			logger.Logger(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create Kafka producer: %s\n", err.Error()))
		}

		closer.AddNamed("Kafka sync producer", func(ctx context.Context) error {
			return d.eventProducer.Close()
		})

		d.eventProducer = producer
	}
	return d.eventProducer
}

func (d *diContainer) JoinTokenService() tokens.JoinTokenService {
	if d.joinTokenService == nil {
		d.joinTokenService = jwtTokens.NewJoinJWTService(config.AppConfig().JWT)
	}
	return d.joinTokenService
}

func (d *diContainer) AccessTokenVerifier() tokens.AccessTokenVerifier {
	if d.accessVerifier == nil {
		d.accessVerifier = jwtTokens.NewAccessJWTVerifier(config.AppConfig().JWT)
	}
	return d.accessVerifier
}

func (d *diContainer) ServiceTokenVerifier() tokens.ServiceTokenVerifier {
	if d.serviceVerifier == nil {
		d.serviceVerifier = jwtTokens.NewServiceJWTService(config.AppConfig().JWT)
	}
	return d.serviceVerifier
}
