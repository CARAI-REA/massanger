package config

import (
	"time"

	webrtcapi "sfu/internal/webrtc"
)

type AppEnvConfig interface {
	Env() string
}

type LoggerConfig interface {
	Level() string
	AsJson() bool
}

type HTTPConfig interface {
	Address() string
	WSPath() string
	PingInterval() time.Duration
	PongWait() time.Duration
	MaxMessageBytes() int64
	MaxMessagesPerSec() int
	AllowedOrigins() []string
	AllowQueryToken() bool
}

type RedisConfig interface {
	Host() string
	Port() string
	Password() string
	ConnectionTimeout() time.Duration
	MaxIdle() int
	IdleTimeout() time.Duration
	AffinityTTL() time.Duration
}

type JWTConfig interface {
	JoinTokenSecretKey() string
	ServiceTokenSecretKey() string
	ServiceName() string
}

type RoomGRPCConfig interface {
	Address() string
	Enabled() bool
}

type KafkaConfig interface {
	Brokers() []string
	RoomEventsTopic() string
	ParticipantEventsTopic() string
	ConsumerGroup() string
	Enabled() bool
}

type MetricsConfig interface {
	Address() string
}

type InstanceConfig interface {
	ID() string
	PublicWSURL() string
}

type WebRTCConfig interface {
	ICEServers() []webrtcapi.ICEServer
	ClientICEServers() []webrtcapi.ICEServer
	UDPPortMin() uint16
	UDPPortMax() uint16
	NAT1To1IPs() []string
	NAT1To1CandidateType() string
	TURNCredential() string
	SimulcastEnabled() bool
	RecordingEnabled() bool
	RecordingDir() string
}
