package room

import "context"

type RoomClient interface {
	AssertCanJoin(ctx context.Context, roomUUID, userUUID string) error
}
