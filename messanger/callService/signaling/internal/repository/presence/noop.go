package presence

import (
	"context"

	"signaling/internal/model"
)

// Noop is an in-memory presence stub used before Redis is wired.
type Noop struct {
	peers map[string]map[string]string // room -> user -> instance
}

func NewNoop() *Noop {
	return &Noop{peers: make(map[string]map[string]string)}
}

func (n *Noop) Register(_ context.Context, roomUUID, userUUID, instanceID string) error {
	if n.peers[roomUUID] == nil {
		n.peers[roomUUID] = make(map[string]string)
	}
	n.peers[roomUUID][userUUID] = instanceID
	return nil
}

func (n *Noop) Unregister(_ context.Context, roomUUID, userUUID string) error {
	if n.peers[roomUUID] != nil {
		delete(n.peers[roomUUID], userUUID)
		if len(n.peers[roomUUID]) == 0 {
			delete(n.peers, roomUUID)
		}
	}
	return nil
}

func (n *Noop) RefreshTTL(context.Context, string, string) error { return nil }

func (n *Noop) ListPeers(_ context.Context, roomUUID string) ([]string, error) {
	m := n.peers[roomUUID]
	out := make([]string, 0, len(m))
	for u := range m {
		out = append(out, u)
	}
	return out, nil
}

func (n *Noop) GetInstance(_ context.Context, roomUUID, userUUID string) (string, bool, error) {
	id, ok := n.peers[roomUUID][userUUID]
	return id, ok, nil
}

func (n *Noop) Publish(context.Context, string, model.Envelope) error { return nil }

func (n *Noop) Subscribe(ctx context.Context, _ func(string, model.Envelope)) error {
	<-ctx.Done()
	return ctx.Err()
}

func (n *Noop) EnsureSubscribed(context.Context, string) error { return nil }
func (n *Noop) ReleaseSubscribe(context.Context, string) error { return nil }
