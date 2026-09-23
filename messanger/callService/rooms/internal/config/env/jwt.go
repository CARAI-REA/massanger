package env

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type jwtEnvConfig struct {
	AuthTokenSecretKey    string        `env:"JWT_SECRET,required"`
	JoinTokenSecretKey    string        `env:"JOIN_TOKEN_SECRET_KEY,required"`
	JoinTokenExpiration   time.Duration `env:"JOIN_TOKEN_EXPIRATION,required"`
	ServiceTokenSecretKey string        `env:"SERVICE_JWT_SECRET,required"`
}

type jwtConfig struct {
	raw jwtEnvConfig
}

func NewJWTConfig() (*jwtConfig, error) {
	var raw jwtEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &jwtConfig{raw: raw}, nil
}

func (cfg *jwtConfig) AuthTokenSecretKey() string {
	return cfg.raw.AuthTokenSecretKey
}

func (cfg *jwtConfig) JoinTokenSecretKey() string {
	return cfg.raw.JoinTokenSecretKey
}

func (cfg *jwtConfig) JoinTokenExpiration() time.Duration {
	return cfg.raw.JoinTokenExpiration
}

func (cfg *jwtConfig) ServiceTokenSecretKey() string {
	return cfg.raw.ServiceTokenSecretKey
}
