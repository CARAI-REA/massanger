package hub_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"signaling/internal/model"
	"signaling/internal/service/hub"
)

type memConn struct {
	mu   sync.Mutex
	msgs []model.Envelope
	closed bool
}

func (c *memConn) Send(_ context.Context, msg model.Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, msg)
	return nil
}

func (c *memConn) Close(int, string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *memConn) RemoteAddr() string { return "test" }

func (c *memConn) Messages() []model.Envelope {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]model.Envelope, len(c.msgs))
	copy(out, c.msgs)
	return out
}

func TestH1_RegisterTwoUsers(t *testing.T) {
	h := hub.New()
	h.Register("r1", "u1", &memConn{})
	h.Register("r1", "u2", &memConn{})
	require.Len(t, h.LocalPeers("r1"), 2)
}

func TestH2_RegisterSameUserEvicts(t *testing.T) {
	h := hub.New()
	c1 := &memConn{}
	c2 := &memConn{}
	require.Nil(t, h.Register("r1", "u1", c1))
	evicted := h.Register("r1", "u1", c2)
	require.Equal(t, c1, evicted)
	got, ok := h.Get("r1", "u1")
	require.True(t, ok)
	require.Equal(t, c2, got)
}

func TestH3_UnregisterRemovesPeer(t *testing.T) {
	h := hub.New()
	c := &memConn{}
	h.Register("r1", "u1", c)
	require.True(t, h.Unregister("r1", "u1", c))
	_, ok := h.Get("r1", "u1")
	require.False(t, ok)
}

func TestH4_BroadcastExcludesSender(t *testing.T) {
	h := hub.New()
	a := &memConn{}
	b := &memConn{}
	h.Register("r1", "u1", a)
	h.Register("r1", "u2", b)
	h.Broadcast(context.Background(), "r1", "u1", model.Envelope{Type: model.TypePeerJoined})
	require.Empty(t, a.Messages())
	require.Len(t, b.Messages(), 1)
}
