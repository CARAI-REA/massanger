package participant

import (
	"context"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *service) IsParticipant(ctx context.Context, req model.IsParticipantRequest) (model.IsParticipantResponse, error) {
	// Without actor (Signaling internal call) — allow lookup.
	// With actor — only self or room owner may query.
	if actor, err := authctx.MustUserID(ctx); err == nil {
		roomRes, err := s.roomRepository.GetRoom(ctx, model.GetRoomRequest{RoomUUID: req.RoomUUID})
		if err != nil {
			return model.IsParticipantResponse{}, err
		}
		room := roomRes.Room
		if actor != req.UserUUID && room.OwnerUUID != actor {
			return model.IsParticipantResponse{}, model.ErrPermissionDenied
		}
	}

	return s.participantRepository.IsParticipant(ctx, req)
}
