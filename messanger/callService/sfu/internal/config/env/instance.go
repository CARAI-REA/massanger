package env

import (
	"os"

	"github.com/caarlos0/env/v11"
)

type instanceEnvConfig struct {
	ID          string `env:"INSTANCE_ID" envDefault:""`
	PublicWSURL string `env:"PUBLIC_WS_URL" envDefault:"ws://localhost:8082/v1/ws"`
}

type instanceConfig struct{ raw instanceEnvConfig }

func NewInstanceConfig() (*instanceConfig, error) {
	var raw instanceEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	if raw.ID == "" {
		host, err := os.Hostname()
		if err != nil || host == "" {
			host = "sfu-unknown"
		}
		raw.ID = host
	}
	return &instanceConfig{raw: raw}, nil
}

func (c *instanceConfig) ID() string          { return c.raw.ID }
func (c *instanceConfig) PublicWSURL() string { return c.raw.PublicWSURL }
