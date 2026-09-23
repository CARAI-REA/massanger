package v1

import (
	"rooms/internal/service"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

type api struct {
	roomsV1.UnimplementedRoomServiceServer

	roomService        service.RoomService
	participantService service.ParticipantService
}

func NewAPI(roomService service.RoomService, participantService service.ParticipantService) *api {
	return &api{
		roomService:        roomService,
		participantService: participantService,
	}
}
