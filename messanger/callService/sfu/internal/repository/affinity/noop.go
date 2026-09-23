package affinity

import (
	"context"
	"sync"
	"time"
)

type Noop struct {
	mu    sync.Mutex
	owner map[string]string // room -> instance|url
}

func NewNoop() *Noop {
	return &Noop{owner: make(map[string]string)}
}

func (n *Noop) Claim(_ context.Context, roomUUID, instanceID, publicURL string, _ time.Duration) (bool, string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	cur, ok := n.owner[roomUUID]
	if !ok {
		n.owner[roomUUID] = encodeOwner(instanceID, publicURL)
		return true, publicURL, nil
	}
	id, url := decodeOwner(cur)
	if id == instanceID {
		return true, url, nil
	}
	return false, url, nil
}

func (n *Noop) Refresh(context.Context, string, string, time.Duration) error { return nil }

func (n *Noop) Release(_ context.Context, roomUUID, instanceID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	cur, ok := n.owner[roomUUID]
	if !ok {
		return nil
	}
	id, _ := decodeOwner(cur)
	if id == instanceID {
		delete(n.owner, roomUUID)
	}
	return nil
}
