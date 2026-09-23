package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type metricsEnvConfig struct {
	Host string `env:"METRICS_HOST" envDefault:"0.0.0.0"`
	Port string `env:"METRICS_PORT,required"`
}

type metricsConfig struct {
	raw metricsEnvConfig
}

func NewMetricsConfig() (*metricsConfig, error) {
	var raw metricsEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &metricsConfig{raw: raw}, nil
}

func (cfg *metricsConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
