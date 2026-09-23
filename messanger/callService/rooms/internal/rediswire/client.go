package rediswire

import (
	"context"
	"net"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	platredis "github.com/CARAI-REA/messanger/callService/platform/pkg/cache/redis"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"

	"rooms/internal/repository"
	"rooms/internal/repository/roomcache"
)

// PoolConfig совпадает по смыслу с space_manufacture/redis/clean_arch/ufo/internal/config/env/redis.go.
type PoolConfig struct {
	Host              string
	Port              string
	Password          string
	ConnectionTimeout time.Duration
	MaxIdle           int
	IdleTimeout       time.Duration
}

// NewPool создаёт redigo.Pool (как в clean_arch DI RedisPool).
func NewPool(cfg PoolConfig) *redigo.Pool {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	return &redigo.Pool{
		MaxIdle:     cfg.MaxIdle,
		IdleTimeout: cfg.IdleTimeout,
		DialContext: func(ctx context.Context) (redigo.Conn, error) {
			opts := []redigo.DialOption{}
			if cfg.Password != "" {
				opts = append(opts, redigo.DialPassword(cfg.Password))
			}
			return redigo.DialContext(ctx, "tcp", addr, opts...)
		},
	}
}

// NewRoomCache оборачивает pool в platform/redis.Client и roomcache.Redis.
func NewRoomCache(pool *redigo.Pool, cfg PoolConfig) repository.RoomCache {
	client := platredis.NewClient(pool, logger.Logger(), cfg.ConnectionTimeout)
	return roomcache.NewRedis(client)
}
