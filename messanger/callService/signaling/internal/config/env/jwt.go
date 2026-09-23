package env

import "github.com/caarlos0/env/v11"

type jwtEnvConfig struct {
	JoinTokenSecretKey    string `env:"JOIN_TOKEN_SECRET_KEY,required"`
	ServiceTokenSecretKey string `env:"SERVICE_JWT_SECRET" envDefault:""`
	ServiceName           string `env:"SERVICE_NAME" envDefault:"signaling"`
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

func (cfg *jwtConfig) JoinTokenSecretKey() string {
	return cfg.raw.JoinTokenSecretKey
}

func (cfg *jwtConfig) ServiceTokenSecretKey() string {
	return cfg.raw.ServiceTokenSecretKey
}

func (cfg *jwtConfig) ServiceName() string {
	return cfg.raw.ServiceName
}
