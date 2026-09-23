package config

import "time"

type LoggerConfig interface {
	Level() string
	AsJson() bool
}

type RoomsGRPCConfig interface {
	Address() string
}

type PostgresConfig interface {
	URI() string
	DatabaseName() string
	MigrationDir() string
	SSLMode() string
}

type RedisConfig interface {
	Host() string
	Port() string
	Password() string
	ConnectionTimeout() time.Duration
	MaxIdle() int
	IdleTimeout() time.Duration
	CacheTTL() time.Duration
}

type KafkaConfig interface {
	Brokers() []string
	RoomEventsTopic() string
	ParticipantEventsTopic() string
}

type JWTConfig interface {
	AuthTokenSecretKey() string
	JoinTokenSecretKey() string
	JoinTokenExpiration() time.Duration
	ServiceTokenSecretKey() string
}

type MetricsConfig interface {
	Address() string
}

type AppEnvConfig interface {
	Env() string
	GRPCMaxRPS() float64
}

type TURNConfig interface {
	SharedSecret() string
	URLs() []string
	TTL() time.Duration
}
