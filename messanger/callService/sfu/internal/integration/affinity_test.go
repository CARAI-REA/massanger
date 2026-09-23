package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"sfu/internal/rediswire"
	"sfu/internal/repository/affinity"
)

func dockerAvailable() (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{Image: "redis:7-alpine"},
		Started:          false,
	})
	return err == nil
}

func TestAffinityRedis(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}
	ctx := context.Background()
	redisC, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Skipf("redis: %v", err)
	}
	t.Cleanup(func() { _ = redisC.Terminate(ctx) })

	host, err := redisC.Host(ctx)
	require.NoError(t, err)
	port, err := redisC.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	pool := rediswire.NewPool(rediswire.PoolConfig{
		Host: host, Port: port.Port(), ConnectionTimeout: 3 * time.Second, MaxIdle: 5, IdleTimeout: time.Minute,
	})
	t.Cleanup(func() { _ = pool.Close() })

	a := affinity.NewRedis(pool)
	ok, url, err := a.Claim(ctx, "room", "i1", "ws://i1", 2*time.Second)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "ws://i1", url)

	ok, url, err = a.Claim(ctx, "room", "i2", "ws://i2", 2*time.Second)
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, "ws://i1", url)

	require.NoError(t, a.Release(ctx, "room", "i1"))
	ok, _, err = a.Claim(ctx, "room", "i2", "ws://i2", 2*time.Second)
	require.NoError(t, err)
	require.True(t, ok)
}
