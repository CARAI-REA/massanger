package v1

import (
	"context"
	"errors"

	"rooms/internal/converter"
	"rooms/internal/model"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *api) GetRoom(ctx context.Context, req *roomsV1.GetRoomRequest) (*roomsV1.GetRoomResponse, error) {
	resp, err := a.roomService.GetRoom(ctx, converter.GetRoomRequestToModel(req))
	if err != nil {
		if errors.Is(err, model.ErrRoomNotFound) {
			return nil, status.Errorf(codes.NotFound, "room with UUID %s not found", req.GetRoomUuid())
		}
		return nil, mapError(err)
	}

	return converter.GetRoomResponseToProto(resp), nil
}
