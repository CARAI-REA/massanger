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
	AffinityTTL       time.Duration `env:"AFFINITY_TTL" envDefault:"30s"`
}

type redisConfig struct{ raw redisEnvConfig }

func NewRedisConfig() (*redisConfig, error) {
	var raw redisEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &redisConfig{raw: raw}, nil
}

func (c *redisConfig) Host() string                     { return c.raw.Host }
func (c *redisConfig) Port() string                     { return c.raw.Port }
func (c *redisConfig) Password() string                 { return c.raw.Password }
func (c *redisConfig) ConnectionTimeout() time.Duration { return c.raw.ConnectionTimeout }
func (c *redisConfig) MaxIdle() int                     { return c.raw.MaxIdle }
func (c *redisConfig) IdleTimeout() time.Duration       { return c.raw.IdleTimeout }
func (c *redisConfig) AffinityTTL() time.Duration       { return c.raw.AffinityTTL }
