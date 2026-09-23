package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) AddParticipant(ctx context.Context, req *roomsV1.AddParticipantRequest) (*roomsV1.AddParticipantResponse, error) {
	resp, err := a.participantService.AddParticipant(ctx, converter.AddParticipantRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.AddParticipantResponseToProto(resp), nil
}
