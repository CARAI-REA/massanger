package env

import (
	"os"

	"github.com/caarlos0/env/v11"
)

type instanceEnvConfig struct {
	ID string `env:"INSTANCE_ID" envDefault:""`
}

type instanceConfig struct {
	raw instanceEnvConfig
}

func NewInstanceConfig() (*instanceConfig, error) {
	var raw instanceEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	if raw.ID == "" {
		host, err := os.Hostname()
		if err != nil || host == "" {
			host = "signaling-unknown"
		}
		raw.ID = host
	}
	return &instanceConfig{raw: raw}, nil
}

func (cfg *instanceConfig) ID() string { return cfg.raw.ID }
