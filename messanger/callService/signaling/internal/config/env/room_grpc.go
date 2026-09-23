package env

import "github.com/caarlos0/env/v11"

type roomGRPCEnvConfig struct {
	Address string `env:"ROOM_GRPC_ADDR" envDefault:"localhost:8080"`
	Enabled bool   `env:"ROOM_GRPC_ENABLED" envDefault:"true"`
}

type roomGRPCConfig struct {
	raw roomGRPCEnvConfig
}

func NewRoomGRPCConfig() (*roomGRPCConfig, error) {
	var raw roomGRPCEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &roomGRPCConfig{raw: raw}, nil
}

func (cfg *roomGRPCConfig) Address() string { return cfg.raw.Address }
func (cfg *roomGRPCConfig) Enabled() bool   { return cfg.raw.Enabled }
