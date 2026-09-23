package room

import (
	"context"
	"fmt"

	"rooms/internal/metrics"
	"rooms/internal/model"
	"rooms/internal/service/access"
	"rooms/internal/service/authctx"
)

func (s *service) CreateRoom(ctx context.Context, req model.CreateRoomRequest) (model.CreateRoomResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.CreateRoomResponse{}, err
	}
	if err := access.ValidateCreateRoom(req); err != nil {
		return model.CreateRoomResponse{}, err
	}
	if actor != req.OwnerUUID {
		return model.CreateRoomResponse{}, model.ErrPermissionDenied
	}
	resp, err := s.roomRepository.CreateRoom(ctx, req)
	if err != nil {
		return model.CreateRoomResponse{}, err
	}
	if err := s.eventProducer.PublishRoomCreated(ctx, resp.Room.RoomUUID, req.OwnerUUID); err != nil {
		return model.CreateRoomResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}
	_ = s.roomCache.SetRoom(ctx, resp.Room, s.cacheTTL)
	metrics.ActiveRooms.Inc()
	return resp, nil
}
