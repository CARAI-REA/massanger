package hub

import (
	"context"
	"sync"

	"signaling/internal/model"
	"signaling/internal/service"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]service.Conn // room -> user -> conn
}

func New() *Hub {
	return &Hub{
		rooms: make(map[string]map[string]service.Conn),
	}
}

func (h *Hub) Register(roomUUID, userUUID string, c service.Conn) (evicted service.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	peers, ok := h.rooms[roomUUID]
	if !ok {
		peers = make(map[string]service.Conn)
		h.rooms[roomUUID] = peers
	}
	if old, exists := peers[userUUID]; exists && old != c {
		evicted = old
	}
	peers[userUUID] = c
	return evicted
}

func (h *Hub) Unregister(roomUUID, userUUID string, c service.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	peers, ok := h.rooms[roomUUID]
	if !ok {
		return false
	}
	cur, exists := peers[userUUID]
	if !exists || (c != nil && cur != c) {
		return false
	}
	delete(peers, userUUID)
	if len(peers) == 0 {
		delete(h.rooms, roomUUID)
	}
	return true
}

func (h *Hub) Get(roomUUID, userUUID string) (service.Conn, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	peers, ok := h.rooms[roomUUID]
	if !ok {
		return nil, false
	}
	c, ok := peers[userUUID]
	return c, ok
}

func (h *Hub) LocalPeers(roomUUID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	peers, ok := h.rooms[roomUUID]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(peers))
	for u := range peers {
		out = append(out, u)
	}
	return out
}

func (h *Hub) Broadcast(ctx context.Context, roomUUID, excludeUser string, msg model.Envelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	peers, ok := h.rooms[roomUUID]
	if !ok {
		return
	}
	for u, c := range peers {
		if u == excludeUser {
			continue
		}
		_ = c.Send(ctx, msg)
	}
}

// syncCloser writes a final message then closes the socket (avoids dropping
// ROOM_CLOSED when the async send queue is closed immediately).
type syncCloser interface {
	SendAndClose(ctx context.Context, msg model.Envelope, code int, reason string) error
}

func (h *Hub) CloseAll(ctx context.Context, roomUUID string, msg model.Envelope) {
	h.mu.Lock()
	peers := h.rooms[roomUUID]
	delete(h.rooms, roomUUID)
	h.mu.Unlock()

	for _, c := range peers {
		if sc, ok := c.(syncCloser); ok {
			_ = sc.SendAndClose(ctx, msg, 1000, "room closed")
			continue
		}
		_ = c.Send(ctx, msg)
		_ = c.Close(1000, "room closed")
	}
}

func (h *Hub) LocalRoomCount(roomUUID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[roomUUID])
}
