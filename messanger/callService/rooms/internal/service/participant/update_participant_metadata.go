package participant

import (
	"context"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *service) UpdateParticipantMetadata(ctx context.Context, req model.UpdateParticipantMetadataRequest) (model.UpdateParticipantMetadataResponse, error) {
	actor, err := authctx.MustUserID(ctx)
	if err != nil {
		return model.UpdateParticipantMetadataResponse{}, err
	}
	if actor != req.UserUUID {
		return model.UpdateParticipantMetadataResponse{}, model.ErrPermissionDenied
	}
	return s.participantRepository.UpdateParticipantMetadata(ctx, req)
}
