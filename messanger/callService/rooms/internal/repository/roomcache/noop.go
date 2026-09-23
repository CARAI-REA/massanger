package roomcache

import (
	"context"
	"time"

	"rooms/internal/model"
	"rooms/internal/repository"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/cache"
)

// Noop заглушка: GetRoom всегда cache miss; ActiveParticipantCount = -1 (пропуск auto-close в тестах).
type Noop struct{}

// NewNoop возвращает заглушку repository.RoomCache для тестов.
func NewNoop() repository.RoomCache {
	return Noop{}
}

var _ repository.RoomCache = Noop{}

func (Noop) SetRoom(context.Context, model.Room, time.Duration) error { return nil }

func (Noop) GetRoom(context.Context, string) (model.Room, error) {
	return model.Room{}, cache.ErrCacheMiss
}

func (Noop) InvalidateRoomInfo(context.Context, string) error { return nil }

func (Noop) PurgeRoom(context.Context, string) error { return nil }

func (Noop) AddActiveParticipant(context.Context, string, string) error { return nil }

func (Noop) RemoveActiveParticipant(context.Context, string, string) error { return nil }

func (Noop) ActiveParticipantCount(context.Context, string) (int64, error) { return -1, nil }
