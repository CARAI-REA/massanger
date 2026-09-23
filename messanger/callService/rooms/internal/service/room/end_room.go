package room

import (
	"context"
	"fmt"

	"rooms/internal/metrics"
	"rooms/internal/model"
	"rooms/internal/service/access"
	"rooms/internal/service/authctx"
)

func (s *service) EndRoom(ctx context.Context, req model.EndRoomRequest) (model.EndRoomResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.EndRoomResponse{}, err
	}
	if err := s.assertActorOwner(actor, req.OwnerUUID); err != nil {
		return model.EndRoomResponse{}, err
	}
	room, err := s.loadRoom(ctx, req.RoomUUID)
	if err != nil {
		return model.EndRoomResponse{}, err
	}
	if room.DeletedAt != nil {
		return model.EndRoomResponse{}, model.ErrRoomNotFound
	}
	if err := s.assertRoomOwner(room, req.OwnerUUID); err != nil {
		return model.EndRoomResponse{}, err
	}
	if !access.RoomActive(room) {
		return model.EndRoomResponse{}, model.ErrRoomInactive
	}
	resp, err := s.roomRepository.EndRoom(ctx, req)
	if err != nil {
		return model.EndRoomResponse{}, err
	}
	if err := s.eventProducer.PublishRoomEnded(ctx, req.RoomUUID); err != nil {
		return model.EndRoomResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}
	_ = s.roomCache.PurgeRoom(ctx, req.RoomUUID)
	metrics.ActiveRooms.Dec()
	return resp, nil
}
