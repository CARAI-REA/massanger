package participant

import (
	"context"
	"fmt"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *service) RefreshJoinToken(ctx context.Context, req model.RefreshJoinTokenRequest) (model.RefreshJoinTokenResponse, error) {
	if req.RoomUUID == "" || req.UserUUID == "" {
		return model.RefreshJoinTokenResponse{}, model.ErrInvalidArgument
	}

	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.RefreshJoinTokenResponse{}, err
	}
	if actor != req.UserUUID {
		return model.RefreshJoinTokenResponse{}, model.ErrPermissionDenied
	}

	partRes, err := s.participantRepository.IsParticipant(ctx, model.IsParticipantRequest{
		RoomUUID: req.RoomUUID,
		UserUUID: req.UserUUID,
	})
	if err != nil {
		return model.RefreshJoinTokenResponse{}, err
	}
	if !partRes.IsActive {
		return model.RefreshJoinTokenResponse{}, model.ErrPermissionDenied
	}

	if s.joinTokenService == nil {
		return model.RefreshJoinTokenResponse{}, fmt.Errorf("join token service is not configured")
	}

	token, err := s.joinTokenService.GenerateJoinToken(ctx, joinInfo{
		userUUID: req.UserUUID,
		roomUUID: req.RoomUUID,
	})
	if err != nil {
		return model.RefreshJoinTokenResponse{}, err
	}

	return model.RefreshJoinTokenResponse{JoinToken: token}, nil
}
