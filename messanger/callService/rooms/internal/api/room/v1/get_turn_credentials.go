package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) GetTURNCredentials(ctx context.Context, req *roomsV1.GetTURNCredentialsRequest) (*roomsV1.GetTURNCredentialsResponse, error) {
	resp, err := a.participantService.GetTURNCredentials(ctx, converter.GetTURNCredentialsRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}
	return converter.GetTURNCredentialsResponseToProto(resp), nil
}
