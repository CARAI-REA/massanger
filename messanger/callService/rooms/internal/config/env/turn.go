package env

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type turnEnvConfig struct {
	SharedSecret string        `env:"TURN_SHARED_SECRET"`
	URLs         string        `env:"TURN_URLS"`
	TTL          time.Duration `env:"TURN_TTL" envDefault:"1h"`
}

type turnConfig struct {
	raw turnEnvConfig
}

func NewTURNConfig() (*turnConfig, error) {
	var raw turnEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &turnConfig{raw: raw}, nil
}

func (cfg *turnConfig) SharedSecret() string {
	return cfg.raw.SharedSecret
}

func (cfg *turnConfig) URLs() []string {
	if strings.TrimSpace(cfg.raw.URLs) == "" {
		return nil
	}
	parts := strings.Split(cfg.raw.URLs, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (cfg *turnConfig) TTL() time.Duration {
	if cfg.raw.TTL <= 0 {
		return time.Hour
	}
	return cfg.raw.TTL
}
