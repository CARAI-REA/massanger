package room

import (
	"context"
	"fmt"

	"rooms/internal/metrics"
	"rooms/internal/model"
	"rooms/internal/service/access"
	"rooms/internal/service/authctx"
)

func (s *service) DeleteRoom(ctx context.Context, req model.DeleteRoomRequest) (model.DeleteRoomResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.DeleteRoomResponse{}, err
	}
	if err := s.assertActorOwner(actor, req.OwnerUUID); err != nil {
		return model.DeleteRoomResponse{}, err
	}
	room, err := s.loadRoom(ctx, req.RoomUUID)
	if err != nil {
		return model.DeleteRoomResponse{}, err
	}
	if room.DeletedAt != nil {
		return model.DeleteRoomResponse{}, model.ErrRoomNotFound
	}
	if err := s.assertRoomOwner(room, req.OwnerUUID); err != nil {
		return model.DeleteRoomResponse{}, err
	}
	if !access.RoomActive(room) {
		return model.DeleteRoomResponse{}, model.ErrRoomInactive
	}
	resp, err := s.roomRepository.DeleteRoom(ctx, req)
	if err != nil {
		return model.DeleteRoomResponse{}, err
	}
	if err := s.eventProducer.PublishRoomDeleted(ctx, req.RoomUUID); err != nil {
		return model.DeleteRoomResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}
	_ = s.roomCache.PurgeRoom(ctx, req.RoomUUID)
	metrics.ActiveRooms.Dec()
	return resp, nil
}
