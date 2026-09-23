package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) UpdateParticipantMetadata(
	ctx context.Context,
	req *roomsV1.UpdateParticipantMetadataRequest,
) (*roomsV1.UpdateParticipantMetadataResponse, error) {
	resp, err := a.participantService.UpdateParticipantMetadata(
		ctx,
		converter.UpdateParticipantMetadataRequestToModel(req),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return converter.UpdateParticipantMetadataResponseToProto(resp), nil
}
