package repository

import (
	"context"

	"signaling/internal/model"
)

type PresenceRepository interface {
	Register(ctx context.Context, roomUUID, userUUID, instanceID string) error
	Unregister(ctx context.Context, roomUUID, userUUID string) error
	RefreshTTL(ctx context.Context, roomUUID, userUUID string) error
	ListPeers(ctx context.Context, roomUUID string) ([]string, error)
	GetInstance(ctx context.Context, roomUUID, userUUID string) (instanceID string, ok bool, err error)
	Publish(ctx context.Context, roomUUID string, msg model.Envelope) error
	Subscribe(ctx context.Context, handler func(roomUUID string, msg model.Envelope)) error
	EnsureSubscribed(ctx context.Context, roomUUID string) error
	ReleaseSubscribe(ctx context.Context, roomUUID string) error
}
