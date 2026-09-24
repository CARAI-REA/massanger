package room

import (
	"context"

	"sfu/internal/model"
)

// JoinCheck is the result of AssertCanJoin.
type JoinCheck struct {
	OK               bool
	RoomStatus       string
	RecordingEnabled bool
}

type RoomClient interface {
	AssertCanJoin(ctx context.Context, roomUUID, userUUID string) (JoinCheck, error)
	GetTURNCredentials(ctx context.Context, roomUUID, userUUID string) ([]model.ICEServer, error)
}
