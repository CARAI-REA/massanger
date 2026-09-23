package room

import (
	"context"
	"fmt"

	"rooms/internal/model"
	"rooms/internal/service/access"
	"rooms/internal/service/authctx"
)

func (s *service) loadRoom(ctx context.Context, roomUUID string) (model.Room, error) {
	res, err := s.roomRepository.GetRoom(ctx, model.GetRoomRequest{RoomUUID: roomUUID})
	if err != nil {
		return model.Room{}, err
	}
	return res.Room, nil
}

func (s *service) assertRoomOwner(room model.Room, ownerUUID string) error {
	if room.OwnerUUID != ownerUUID {
		return model.ErrPermissionDenied
	}
	return nil
}

func (s *service) assertActorOwner(actor, ownerInRequest string) error {
	if actor != ownerInRequest {
		return model.ErrPermissionDenied
	}
	return nil
}

func (s *service) UpdateRoom(ctx context.Context, req model.UpdateRoomRequest) (model.UpdateRoomResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.UpdateRoomResponse{}, err
	}
	if err := s.assertActorOwner(actor, req.OwnerUUID); err != nil {
		return model.UpdateRoomResponse{}, err
	}
	room, err := s.loadRoom(ctx, req.RoomUUID)
	if err != nil {
		return model.UpdateRoomResponse{}, err
	}
	if room.DeletedAt != nil {
		return model.UpdateRoomResponse{}, model.ErrRoomNotFound
	}
	if err := s.assertRoomOwner(room, req.OwnerUUID); err != nil {
		return model.UpdateRoomResponse{}, err
	}
	if !access.RoomActive(room) {
		return model.UpdateRoomResponse{}, model.ErrRoomInactive
	}
	resp, err := s.roomRepository.UpdateRoom(ctx, req)
	if err != nil {
		return model.UpdateRoomResponse{}, err
	}
	if err := s.eventProducer.PublishRoomUpdated(ctx, req.RoomUUID); err != nil {
		return model.UpdateRoomResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}
	_ = s.roomCache.InvalidateRoomInfo(ctx, req.RoomUUID)
	return resp, nil
}
