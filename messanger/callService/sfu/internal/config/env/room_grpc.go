package env

import "github.com/caarlos0/env/v11"

type roomGRPCEnvConfig struct {
	Address string `env:"ROOM_GRPC_ADDR" envDefault:"localhost:8080"`
	Enabled bool   `env:"ROOM_GRPC_ENABLED" envDefault:"true"`
}

type roomGRPCConfig struct{ raw roomGRPCEnvConfig }

func NewRoomGRPCConfig() (*roomGRPCConfig, error) {
	var raw roomGRPCEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &roomGRPCConfig{raw: raw}, nil
}

func (c *roomGRPCConfig) Address() string { return c.raw.Address }
func (c *roomGRPCConfig) Enabled() bool   { return c.raw.Enabled }
