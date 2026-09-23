package env

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type redisEnvConfig struct {
	Host              string        `env:"REDIS_HOST,required"`
	Port              string        `env:"REDIS_PORT,required"`
	Password          string        `env:"REDIS_PASSWORD" envDefault:""`
	ConnectionTimeout time.Duration `env:"REDIS_CONNECTION_TIMEOUT,required"`
	MaxIdle           int           `env:"REDIS_MAX_IDLE,required"`
	IdleTimeout       time.Duration `env:"REDIS_IDLE_TIMEOUT,required"`
	PresenceTTL       time.Duration `env:"PRESENCE_TTL" envDefault:"30s"`
}

type redisConfig struct {
	raw redisEnvConfig
}

func NewRedisConfig() (*redisConfig, error) {
	var raw redisEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &redisConfig{raw: raw}, nil
}

func (cfg *redisConfig) Host() string                     { return cfg.raw.Host }
func (cfg *redisConfig) Port() string                     { return cfg.raw.Port }
func (cfg *redisConfig) Password() string                 { return cfg.raw.Password }
func (cfg *redisConfig) ConnectionTimeout() time.Duration { return cfg.raw.ConnectionTimeout }
func (cfg *redisConfig) MaxIdle() int                     { return cfg.raw.MaxIdle }
func (cfg *redisConfig) IdleTimeout() time.Duration       { return cfg.raw.IdleTimeout }
func (cfg *redisConfig) PresenceTTL() time.Duration       { return cfg.raw.PresenceTTL }
