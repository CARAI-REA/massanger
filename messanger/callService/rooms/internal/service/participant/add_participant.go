package participant

import (
	"context"
	"encoding/json"
	"fmt"

	"rooms/internal/metrics"
	"rooms/internal/model"
	"rooms/internal/service/access"
	"rooms/internal/service/authctx"
)

func (s *service) AddParticipant(ctx context.Context, req model.AddParticipantRequest) (model.AddParticipantResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.AddParticipantResponse{}, err
	}
	if actor != req.UserUUID {
		return model.AddParticipantResponse{}, model.ErrPermissionDenied
	}

	roomRes, err := s.roomRepository.GetRoom(ctx, model.GetRoomRequest{RoomUUID: req.RoomUUID})
	if err != nil {
		return model.AddParticipantResponse{}, err
	}
	room := roomRes.Room
	if room.DeletedAt != nil {
		return model.AddParticipantResponse{}, model.ErrRoomNotFound
	}
	if !access.RoomActive(room) {
		return model.AddParticipantResponse{}, model.ErrRoomInactive
	}

	partRes, err := s.participantRepository.GetParticipant(ctx, model.GetParticipantsRequest{
		RoomUUID:   req.RoomUUID,
		OnlyActive: true,
	})
	if err != nil {
		return model.AddParticipantResponse{}, err
	}
	if partRes.ActiveCount >= room.RoomSettings.MaxParticipants {
		return model.AddParticipantResponse{}, model.ErrResourceExhausted
	}

	if !access.UserMayJoinPrivate(room.RoomSettings, room, req.UserUUID) {
		return model.AddParticipantResponse{}, model.ErrPermissionDenied
	}

	isRes, err := s.participantRepository.IsParticipant(ctx, model.IsParticipantRequest{
		RoomUUID: req.RoomUUID,
		UserUUID: req.UserUUID,
	})
	if err != nil {
		return model.AddParticipantResponse{}, err
	}
	if isRes.IsActive {
		return model.AddParticipantResponse{}, model.ErrParticipantConflict
	}

	resp, err := s.participantRepository.AddParticipant(ctx, req)
	if err != nil {
		return model.AddParticipantResponse{}, err
	}
	metaJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return model.AddParticipantResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}
	if err := s.eventProducer.PublishParticipantJoined(ctx, req.RoomUUID, req.UserUUID, metaJSON); err != nil {
		return model.AddParticipantResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}
	_ = s.roomCache.AddActiveParticipant(ctx, req.RoomUUID, req.UserUUID)
	_ = s.roomCache.InvalidateRoomInfo(ctx, req.RoomUUID)
	metrics.ActiveParticipants.Inc()
	return resp, nil
}
