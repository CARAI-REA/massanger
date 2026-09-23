package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) IsParticipant(ctx context.Context, req *roomsV1.IsParticipantRequest) (*roomsV1.IsParticipantResponse, error) {
	resp, err := a.participantService.IsParticipant(ctx, converter.IsParticipantRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.IsParticipantResponseToProto(resp), nil
}
