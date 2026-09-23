package participant

import (
	"context"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *service) GetParticipant(ctx context.Context, req model.GetParticipantsRequest) (model.GetParticipantsResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.GetParticipantsResponse{}, err
	}

	roomRes, err := s.roomRepository.GetRoom(ctx, model.GetRoomRequest{RoomUUID: req.RoomUUID})
	if err != nil {
		return model.GetParticipantsResponse{}, err
	}
	room := roomRes.Room

	if room.OwnerUUID != actor {
		isRes, err := s.participantRepository.IsParticipant(ctx, model.IsParticipantRequest{
			RoomUUID: req.RoomUUID,
			UserUUID: actor,
		})
		if err != nil {
			return model.GetParticipantsResponse{}, err
		}
		if !isRes.IsActive {
			return model.GetParticipantsResponse{}, model.ErrPermissionDenied
		}
	}

	return s.participantRepository.GetParticipant(ctx, req)
}
