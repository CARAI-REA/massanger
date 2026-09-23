package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) AssertCanJoin(ctx context.Context, req *roomsV1.AssertCanJoinRequest) (*roomsV1.AssertCanJoinResponse, error) {
	resp, err := a.participantService.AssertCanJoin(ctx, converter.AssertCanJoinRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}
	return converter.AssertCanJoinResponseToProto(resp), nil
}
