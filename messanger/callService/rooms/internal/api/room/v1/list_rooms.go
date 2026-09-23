package v1

import (
	"context"

	"rooms/internal/converter"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (a *api) ListRooms(ctx context.Context, req *roomsV1.ListRoomsRequest) (*roomsV1.ListRoomsResponse, error) {
	resp, err := a.roomService.ListRooms(ctx, converter.ListRoomRequestToModel(req))
	if err != nil {
		return nil, mapError(err)
	}

	return converter.ListRoomsResponseToProto(resp), nil
}
