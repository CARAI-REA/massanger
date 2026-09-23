package env

import "github.com/caarlos0/env/v11"

type jwtEnvConfig struct {
	JoinTokenSecretKey    string `env:"JOIN_TOKEN_SECRET_KEY,required"`
	ServiceTokenSecretKey string `env:"SERVICE_JWT_SECRET" envDefault:""`
	ServiceName           string `env:"SERVICE_NAME" envDefault:"sfu"`
}

type jwtConfig struct{ raw jwtEnvConfig }

func NewJWTConfig() (*jwtConfig, error) {
	var raw jwtEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &jwtConfig{raw: raw}, nil
}

func (c *jwtConfig) JoinTokenSecretKey() string    { return c.raw.JoinTokenSecretKey }
func (c *jwtConfig) ServiceTokenSecretKey() string { return c.raw.ServiceTokenSecretKey }
func (c *jwtConfig) ServiceName() string           { return c.raw.ServiceName }
