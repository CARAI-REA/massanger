package app

import (
	"context"
	"fmt"
	"net/http"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/closer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	jwtTokens "github.com/CARAI-REA/messanger/callService/platform/pkg/tokens/jwt"

	wsv1 "signaling/internal/api/ws/v1"
	"signaling/internal/client/room"
	roomgrpc "signaling/internal/client/room/grpc"
	"signaling/internal/config"
	"signaling/internal/kafka"
	"signaling/internal/model"
	"signaling/internal/rediswire"
	"signaling/internal/repository"
	"signaling/internal/repository/presence"
	"signaling/internal/service"
	"signaling/internal/service/hub"
	"signaling/internal/service/session"
)

type diContainer struct {
	hub            *hub.Hub
	sessionService service.SessionService
	presence       repository.PresenceRepository
	redisPool      *redigo.Pool
	joinVerifier   tokens.JoinTokenVerifier
	serviceTokens  tokens.ServiceTokenService
	roomClient     room.RoomClient
	roomGRPC       *roomgrpc.Client
	kafkaConsumer  *kafka.Consumer
	httpMux        http.Handler
}

func NewDiContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) Hub() *hub.Hub {
	if d.hub == nil {
		d.hub = hub.New()
	}
	return d.hub
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

func (d *diContainer) Presence(ctx context.Context) repository.PresenceRepository {
	if d.presence == nil {
		d.presence = presence.NewRedis(d.RedisPool(), config.AppConfig().Redis.PresenceTTL())
	}
	return d.presence
}

func (d *diContainer) JoinVerifier() tokens.JoinTokenVerifier {
	if d.joinVerifier == nil {
		d.joinVerifier = jwtTokens.NewJoinJWTVerifier(config.AppConfig().JWT)
	}
	return d.joinVerifier
}

func (d *diContainer) ServiceTokens() tokens.ServiceTokenService {
	if d.serviceTokens == nil {
		d.serviceTokens = jwtTokens.NewServiceJWTService(config.AppConfig().JWT)
	}
	return d.serviceTokens
}

func (d *diContainer) RoomClient(ctx context.Context) room.RoomClient {
	if d.roomClient == nil {
		if !config.AppConfig().RoomGRPC.Enabled() {
			d.roomClient = room.NewNoop()
			return d.roomClient
		}
		client, err := roomgrpc.New(
			config.AppConfig().RoomGRPC.Address(),
			d.ServiceTokens(),
			config.AppConfig().JWT.ServiceName(),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create room gRPC client: %v", err))
		}
		closer.AddNamed("Room gRPC client", func(ctx context.Context) error {
			return client.Close()
		})
		d.roomGRPC = client
		d.roomClient = client
	}
	return d.roomClient
}

func (d *diContainer) SessionService(ctx context.Context) service.SessionService {
	if d.sessionService == nil {
		d.sessionService = session.New(
			d.Hub(),
			d.Presence(ctx),
			d.RoomClient(ctx),
			config.AppConfig().Instance.ID(),
			config.AppConfig().RoomGRPC.Enabled(),
			config.AppConfig().SFU.Enabled(),
			config.AppConfig().SFU.PublicURL(),
		)
	}
	return d.sessionService
}

func (d *diContainer) KafkaConsumer(ctx context.Context) *kafka.Consumer {
	if !config.AppConfig().Kafka.Enabled() {
		return nil
	}
	if len(config.AppConfig().Kafka.Brokers()) == 0 {
		return nil
	}
	if d.kafkaConsumer == nil {
		topics := []string{
			config.AppConfig().Kafka.RoomEventsTopic(),
			config.AppConfig().Kafka.ParticipantEventsTopic(),
		}
		c, err := kafka.NewConsumer(
			config.AppConfig().Kafka.Brokers(),
			config.AppConfig().Kafka.ConsumerGroup(),
			topics,
			d.SessionService(ctx),
		)
		if err != nil {
			// Do not block HTTP/WS startup if Kafka is temporarily unavailable.
			fmt.Printf("warning: kafka consumer disabled: %v\n", err)
			return nil
		}
		closer.AddNamed("Kafka consumer", func(ctx context.Context) error {
			return d.kafkaConsumer.Close()
		})
		d.kafkaConsumer = c
	}
	return d.kafkaConsumer
}

func (d *diContainer) StartPresenceSubscribe(ctx context.Context) {
	pres := d.Presence(ctx)
	sess, ok := d.SessionService(ctx).(*session.Service)
	if !ok {
		return
	}
	handler := func(roomUUID string, msg model.Envelope) {
		sess.OnRemoteMessage(ctx, roomUUID, msg)
	}
	if r, ok := pres.(*presence.Redis); ok {
		r.SetHandler(handler)
	}
	go func() {
		_ = pres.Subscribe(ctx, handler)
	}()
}

func (d *diContainer) HTTPMux(ctx context.Context) http.Handler {
	if d.httpMux != nil {
		return d.httpMux
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		conn, err := d.RedisPool().GetContext(r.Context())
		if err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
		defer conn.Close()
		if _, err := conn.Do("PING"); err != nil {
			http.Error(w, "redis ping failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	cfg := config.AppConfig().HTTP
	wsHandler := wsv1.NewHandler(
		d.SessionService(ctx),
		d.JoinVerifier(),
		cfg.WSPath(),
		cfg.AllowedOrigins(),
		cfg.PingInterval(),
		cfg.PongWait(),
		cfg.MaxMessageBytes(),
		cfg.MaxMessagesPerSec(),
		cfg.AllowQueryToken(),
	)
	mux.Handle(cfg.WSPath(), wsHandler)

	d.httpMux = mux
	return d.httpMux
}
