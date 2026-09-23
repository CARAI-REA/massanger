package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) CreateRoom(ctx context.Context, req *roomsV1.CreateRoomRequest) (*roomsV1.CreateRoomResponse, error) {
	resp, err := a.roomService.CreateRoom(ctx, converter.CreateRoomRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.CreateRoomResponseToProto(resp), nil
}
