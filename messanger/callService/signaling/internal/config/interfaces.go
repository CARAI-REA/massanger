package config

import "time"

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
	PresenceTTL() time.Duration
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
}

type SFUConfig interface {
	Enabled() bool
	PublicURL() string
}
