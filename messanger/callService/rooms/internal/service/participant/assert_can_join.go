package participant

import (
	"context"

	"rooms/internal/model"
)

func (s *service) AssertCanJoin(ctx context.Context, req model.AssertCanJoinRequest) (model.AssertCanJoinResponse, error) {
	if req.RoomUUID == "" || req.UserUUID == "" {
		return model.AssertCanJoinResponse{}, model.ErrInvalidArgument
	}

	roomRes, err := s.roomRepository.GetRoom(ctx, model.GetRoomRequest{RoomUUID: req.RoomUUID})
	if err != nil {
		return model.AssertCanJoinResponse{}, err
	}

	partRes, err := s.participantRepository.IsParticipant(ctx, model.IsParticipantRequest{
		RoomUUID: req.RoomUUID,
		UserUUID: req.UserUUID,
	})
	if err != nil {
		return model.AssertCanJoinResponse{}, err
	}

	status := roomRes.Room.Status
	ok := status == "active" && partRes.IsActive
	return model.AssertCanJoinResponse{
		Ok:         ok,
		RoomStatus: status,
	}, nil
}
