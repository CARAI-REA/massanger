package room

import (
	"context"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *service) ListRooms(ctx context.Context, req model.ListRoomRequest) (model.ListRoomResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.ListRoomResponse{}, err
	}
	if actor != req.OwnerUUID {
		return model.ListRoomResponse{}, model.ErrPermissionDenied
	}
	return s.roomRepository.ListRooms(ctx, req)
}
