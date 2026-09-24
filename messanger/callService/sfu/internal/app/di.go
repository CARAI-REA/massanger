package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/closer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/prodguard"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"
	jwtTokens "github.com/CARAI-REA/messanger/callService/platform/pkg/tokens/jwt"

	wsv1 "sfu/internal/api/ws/v1"
	"sfu/internal/client/room"
	roomgrpc "sfu/internal/client/room/grpc"
	"sfu/internal/config"
	"sfu/internal/kafka"
	"sfu/internal/rediswire"
	"sfu/internal/repository"
	"sfu/internal/repository/affinity"
	"sfu/internal/service"
	"sfu/internal/service/roommgr"
	"sfu/internal/service/session"
	webrtcapi "sfu/internal/webrtc"
)

type diContainer struct {
	roomManager    *roommgr.Manager
	sessionService service.SessionService
	affinity       repository.AffinityRepository
	redisPool      *redigo.Pool
	joinVerifier   tokens.JoinTokenVerifier
	serviceTokens  tokens.ServiceTokenService
	roomClient     room.RoomClient
	kafkaConsumer  *kafka.Consumer
	webrtcFactory  *webrtcapi.Factory
	httpMux        http.Handler
}

func NewDiContainer() *diContainer { return &diContainer{} }

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

func (d *diContainer) Affinity(_ context.Context) repository.AffinityRepository {
	if d.affinity == nil {
		d.affinity = affinity.NewRedis(d.RedisPool())
	}
	return d.affinity
}

func (d *diContainer) WebRTCFactory() *webrtcapi.Factory {
	if d.webrtcFactory == nil {
		wc := config.AppConfig().WebRTC
		f, err := webrtcapi.NewFactory(webrtcapi.Config{
			ICEServers:           wc.ICEServers(),
			UDPPortMin:           wc.UDPPortMin(),
			UDPPortMax:           wc.UDPPortMax(),
			NAT1To1IPs:           wc.NAT1To1IPs(),
			NAT1To1CandidateType: wc.NAT1To1CandidateType(),
		})
		if err != nil {
			panic(fmt.Sprintf("webrtc factory: %v", err))
		}
		d.webrtcFactory = f
	}
	return d.webrtcFactory
}

func (d *diContainer) RoomManager() *roommgr.Manager {
	if d.roomManager == nil {
		wc := config.AppConfig().WebRTC
		d.roomManager = roommgr.NewManager(d.WebRTCFactory(), wc.ClientICEServers(), roommgr.Options{
			SimulcastEnabled: wc.SimulcastEnabled(),
			RecordingEnabled: wc.RecordingEnabled(),
			RecordingDir:     wc.RecordingDir(),
		})
		closer.AddNamed("SFU room manager", func(ctx context.Context) error {
			rooms := d.roomManager.RoomUUIDs()
			d.roomManager.CloseAll()
			aff := d.Affinity(ctx)
			instanceID := config.AppConfig().Instance.ID()
			for _, roomUUID := range rooms {
				_ = aff.Release(ctx, roomUUID, instanceID)
			}
			return nil
		})
	}
	return d.roomManager
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

func (d *diContainer) RoomClient(_ context.Context) room.RoomClient {
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
			panic(fmt.Sprintf("room grpc: %v", err))
		}
		closer.AddNamed("Room gRPC client", func(ctx context.Context) error {
			return client.Close()
		})
		d.roomClient = client
	}
	return d.roomClient
}

func (d *diContainer) SessionService(ctx context.Context) service.SessionService {
	if d.sessionService == nil {
		d.sessionService = session.New(
			d.RoomManager(),
			d.Affinity(ctx),
			d.RoomClient(ctx),
			config.AppConfig().Instance.ID(),
			config.AppConfig().Instance.PublicWSURL(),
			config.AppConfig().Redis.AffinityTTL(),
			config.AppConfig().RoomGRPC.Enabled(),
			prodguard.IsProduction(config.AppConfig().AppEnv.Env()) && config.AppConfig().RoomGRPC.Enabled(),
		)
	}
	return d.sessionService
}

func (d *diContainer) KafkaConsumer(ctx context.Context) *kafka.Consumer {
	if !config.AppConfig().Kafka.Enabled() || len(config.AppConfig().Kafka.Brokers()) == 0 {
		return nil
	}
	if d.kafkaConsumer == nil {
		c, err := kafka.NewConsumer(
			config.AppConfig().Kafka.Brokers(),
			config.AppConfig().Kafka.ConsumerGroup(),
			[]string{
				config.AppConfig().Kafka.RoomEventsTopic(),
				config.AppConfig().Kafka.ParticipantEventsTopic(),
			},
			d.SessionService(ctx),
		)
		if err != nil {
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
		if d.SessionService(ctx).IsDraining() {
			http.Error(w, "draining", http.StatusServiceUnavailable)
			return
		}
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
	mux.HandleFunc("/internal/drain", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		if host != "127.0.0.1" && host != "::1" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		d.SessionService(ctx).BeginDrain()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("draining"))
	})
	cfg := config.AppConfig().HTTP
	mux.Handle(cfg.WSPath(), wsv1.NewHandler(
		d.SessionService(ctx),
		d.JoinVerifier(),
		cfg.WSPath(),
		cfg.AllowedOrigins(),
		cfg.PingInterval(),
		cfg.PongWait(),
		cfg.MaxMessageBytes(),
		cfg.MaxMessagesPerSec(),
		cfg.AllowQueryToken(),
	))
	d.httpMux = mux
	return d.httpMux
}
