package env

import (
	"net"

	"github.com/caarlos0/env/v11"
)

type roomsGRPCEnvConfig struct {
	Host string `env:"GRPC_HOST,required"`
	Port string `env:"GRPC_PORT,required"`
}

type roomsGRPCConfig struct {
	raw roomsGRPCEnvConfig
}

func NewRoomsGRPCConfig() (*roomsGRPCConfig, error) {
	var raw roomsGRPCEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &roomsGRPCConfig{raw: raw}, nil
}

func (cfg *roomsGRPCConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}
