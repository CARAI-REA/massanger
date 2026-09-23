package env

import (
	"github.com/caarlos0/env/v11"
)

type appEnvConfig struct {
	Env        string  `env:"APP_ENV" envDefault:"development"`
	GRPCMaxRPS float64 `env:"GRPC_MAX_RPS" envDefault:"0"`
}

type appConfig struct {
	raw appEnvConfig
}

func NewAppConfig() (*appConfig, error) {
	var raw appEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &appConfig{raw: raw}, nil
}

func (cfg *appConfig) Env() string {
	return cfg.raw.Env
}

func (cfg *appConfig) GRPCMaxRPS() float64 {
	return cfg.raw.GRPCMaxRPS
}
