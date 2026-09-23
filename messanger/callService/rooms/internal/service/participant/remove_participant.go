package participant

import (
	"context"
	"fmt"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"

	"rooms/internal/metrics"
	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *service) RemoveParticipant(ctx context.Context, req model.RemoveParticipantRequest) (model.RemoveParticipantResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.RemoveParticipantResponse{}, err
	}

	removedBy := req.RemovedBy
	if removedBy == "" {
		removedBy = req.UserUUID
	}

	roomRes, err := s.roomRepository.GetRoom(ctx, model.GetRoomRequest{RoomUUID: req.RoomUUID})
	if err != nil {
		return model.RemoveParticipantResponse{}, err
	}
	room := roomRes.Room

	if removedBy == req.UserUUID {
		if actor != req.UserUUID {
			return model.RemoveParticipantResponse{}, model.ErrPermissionDenied
		}
	} else {
		if actor != removedBy {
			return model.RemoveParticipantResponse{}, model.ErrPermissionDenied
		}
		if room.OwnerUUID != removedBy {
			return model.RemoveParticipantResponse{}, model.ErrPermissionDenied
		}
	}

	outReq := model.RemoveParticipantRequest{
		RoomUUID:  req.RoomUUID,
		UserUUID:  req.UserUUID,
		RemovedBy: removedBy,
	}

	resp, err := s.participantRepository.RemoveParticipant(ctx, outReq)
	if err != nil {
		return model.RemoveParticipantResponse{}, err
	}
	if err := s.eventProducer.PublishParticipantLeft(ctx, req.RoomUUID, req.UserUUID); err != nil {
		return model.RemoveParticipantResponse{}, fmt.Errorf("%w: %w", model.ErrKafkaPublish, err)
	}

	_ = s.roomCache.RemoveActiveParticipant(ctx, req.RoomUUID, req.UserUUID)
	_ = s.roomCache.InvalidateRoomInfo(ctx, req.RoomUUID)
	metrics.ActiveParticipants.Dec()

	cnt, err := s.roomCache.ActiveParticipantCount(ctx, req.RoomUUID)
	if err == nil && cnt == 0 && room.RoomSettings.AutoClose {
		_, endErr := s.roomRepository.EndRoom(ctx, model.EndRoomRequest{
			RoomUUID:  req.RoomUUID,
			OwnerUUID: room.OwnerUUID,
		})
		if endErr != nil {
			logger.Warn(ctx, "auto-close EndRoom failed", zap.Error(endErr))
		} else {
			if pubErr := s.eventProducer.PublishRoomEnded(ctx, req.RoomUUID); pubErr != nil {
				logger.Warn(ctx, "auto-close PublishRoomEnded failed", zap.Error(pubErr))
			}
			_ = s.roomCache.PurgeRoom(ctx, req.RoomUUID)
			metrics.ActiveRooms.Dec()
		}
	}

	return resp, nil
}
