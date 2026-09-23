package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) GetParticipants(ctx context.Context, req *roomsV1.GetParticipantsRequest) (*roomsV1.GetParticipantsResponse, error) {
	resp, err := a.participantService.GetParticipant(ctx, converter.GetParticipantRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.GetParticipantsResponseToProto(resp), nil
}
