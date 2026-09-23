package env

import "github.com/caarlos0/env/v11"

type sfuEnvConfig struct {
	Enabled   bool   `env:"SFU_ENABLED" envDefault:"false"`
	PublicURL string `env:"SFU_PUBLIC_URL" envDefault:""`
}

type sfuConfig struct{ raw sfuEnvConfig }

func NewSFUConfig() (*sfuConfig, error) {
	var raw sfuEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &sfuConfig{raw: raw}, nil
}

func (c *sfuConfig) Enabled() bool   { return c.raw.Enabled }
func (c *sfuConfig) PublicURL() string { return c.raw.PublicURL }
