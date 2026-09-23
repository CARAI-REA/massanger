package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) EndRoom(ctx context.Context, req *roomsV1.EndRoomRequest) (*roomsV1.EndRoomResponse, error) {
	resp, err := a.roomService.EndRoom(ctx, converter.EndRoomRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.EndRoomResponseToProto(resp), nil
}
