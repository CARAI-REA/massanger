package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) UpdateRoom(ctx context.Context, req *roomsV1.UpdateRoomRequest) (*roomsV1.UpdateRoomResponse, error) {
	resp, err := a.roomService.UpdateRoom(ctx, converter.UpdateRoomRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.UpdateRoomResponseToProto(resp), nil
}
