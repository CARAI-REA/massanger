package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"signaling/internal/model"
	"signaling/internal/rediswire"
	"signaling/internal/repository/presence"
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

func TestI1_I2_I3_RedisPresence(t *testing.T) {
	if !dockerAvailable() {
		t.Skip("docker not available")
	}

	ctx := context.Background()
	redisC, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Skipf("redis container: %v", err)
	}
	t.Cleanup(func() { _ = redisC.Terminate(ctx) })

	host, err := redisC.Host(ctx)
	require.NoError(t, err)
	port, err := redisC.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	poolA := rediswire.NewPool(rediswire.PoolConfig{
		Host: host, Port: port.Port(), ConnectionTimeout: 3 * time.Second, MaxIdle: 5, IdleTimeout: time.Minute,
	})
	poolB := rediswire.NewPool(rediswire.PoolConfig{
		Host: host, Port: port.Port(), ConnectionTimeout: 3 * time.Second, MaxIdle: 5, IdleTimeout: time.Minute,
	})
	t.Cleanup(func() { _ = poolA.Close(); _ = poolB.Close() })

	ttl := 2 * time.Second
	a := presence.NewRedis(poolA, ttl)
	b := presence.NewRedis(poolB, ttl)

	// I1 Register visible via ListPeers on second repo
	require.NoError(t, a.Register(ctx, "room", "user-a", "inst-a"))
	peers, err := b.ListPeers(ctx, "room")
	require.NoError(t, err)
	require.Contains(t, peers, "user-a")

	// I2 Publish from A received on B
	var wg sync.WaitGroup
	wg.Add(1)
	var got model.Envelope
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	require.NoError(t, b.EnsureSubscribed(subCtx, "room"))
	go func() {
		_ = b.Subscribe(subCtx, func(_ string, msg model.Envelope) {
			got = msg
			wg.Done()
		})
	}()
	// give subscribe loop a moment; handler is set in Subscribe
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, a.Publish(ctx, "room", model.Envelope{Type: model.TypeOffer, ToUserUUID: "user-b"}))
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
		require.Equal(t, model.TypeOffer, got.Type)
	case <-time.After(3 * time.Second):
		t.Fatal("pubsub timeout")
	}
	cancel()

	// I3 TTL expire
	require.NoError(t, a.Register(ctx, "room2", "user-x", "inst-a"))
	time.Sleep(ttl + time.Second)
	_, ok, err := b.GetInstance(ctx, "room2", "user-x")
	require.NoError(t, err)
	require.False(t, ok)
}
