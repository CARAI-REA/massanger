package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) RefreshJoinToken(ctx context.Context, req *roomsV1.RefreshJoinTokenRequest) (*roomsV1.RefreshJoinTokenResponse, error) {
	resp, err := a.participantService.RefreshJoinToken(ctx, converter.RefreshJoinTokenRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}
	return converter.RefreshJoinTokenResponseToProto(resp), nil
}
