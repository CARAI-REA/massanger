package rediswire

import (
	"context"
	"net"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

type PoolConfig struct {
	Host              string
	Port              string
	Password          string
	ConnectionTimeout time.Duration
	MaxIdle           int
	IdleTimeout       time.Duration
}

func NewPool(cfg PoolConfig) *redigo.Pool {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	return &redigo.Pool{
		MaxIdle:     cfg.MaxIdle,
		IdleTimeout: cfg.IdleTimeout,
		DialContext: func(ctx context.Context) (redigo.Conn, error) {
			opts := []redigo.DialOption{
				redigo.DialConnectTimeout(cfg.ConnectionTimeout),
			}
			if cfg.Password != "" {
				opts = append(opts, redigo.DialPassword(cfg.Password))
			}
			return redigo.DialContext(ctx, "tcp", addr, opts...)
		},
	}
}
