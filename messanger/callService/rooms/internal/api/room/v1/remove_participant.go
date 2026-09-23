package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) RemoveParticipant(ctx context.Context, req *roomsV1.RemoveParticipantRequest) (*roomsV1.RemoveParticipantResponse, error) {
	resp, err := a.participantService.RemoveParticipant(ctx, converter.RemoveParticipantRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.RemoveParticipantResponseToProto(resp), nil
}
