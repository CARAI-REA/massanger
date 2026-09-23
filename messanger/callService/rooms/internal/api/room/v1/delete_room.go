package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) DeleteRoom(ctx context.Context, req *roomsV1.DeleteRoomRequest) (*roomsV1.DeleteRoomResponse, error) {
	resp, err := a.roomService.DeleteRoom(ctx, converter.DeleteRoomRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.DeleteRoomResponseToProto(resp), nil
}
