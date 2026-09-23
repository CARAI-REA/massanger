package room

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestGetRoomSuccess() {
	var (
		actor     = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		name      = gofakeit.Word()
		ownerUUID = gofakeit.UUID()
		status    = gofakeit.Word()
		createdAt = gofakeit.Date()
		updatedAt = gofakeit.Date()

		getRoomRequest = model.GetRoomRequest{
			RoomUUID: roomUUID,
		}

		room = model.Room{
			RoomUUID:  roomUUID,
			Name:      name,
			OwnerUUID: ownerUUID,
			CratedAt:  lo.ToPtr(createdAt),
			UpdatedAt: lo.ToPtr(updatedAt),
			RoomSettings: model.RoomSettings{
				MaxParticipants: gofakeit.Int32(),
				AllowedUsers:    []string{gofakeit.UUID()},
				Quality:         gofakeit.Word(),
				AutoClose:       gofakeit.Bool(),
			},
			Status: status,
		}
	)

	ctx := authctx.WithUserID(context.Background(), actor)
	s.roomRepository.On("GetRoom", ctx, getRoomRequest).Return(model.GetRoomResponse{Room: room}, nil)

	res, err := s.service.GetRoom(ctx, getRoomRequest)
	s.NoError(err)
	s.Equal(room, res.Room)
}

func (s *ServiceSuite) TestGetRoomError() {
	var (
		actor    = gofakeit.UUID()
		roomUUID = gofakeit.UUID()
		repoErr  = gofakeit.Error()

		getRoomRequest = model.GetRoomRequest{
			RoomUUID: roomUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), actor)
	s.roomRepository.On("GetRoom", ctx, getRoomRequest).Return(model.GetRoomResponse{}, repoErr)

	res, err := s.service.GetRoom(ctx, getRoomRequest)

	s.Require().Error(err)
	s.Require().ErrorIs(err, repoErr)
	s.Require().Empty(res)
}

func (s *ServiceSuite) TestGetRoomWithoutAuthAllowed() {
	roomUUID := gofakeit.UUID()
	room := model.Room{RoomUUID: roomUUID, Name: gofakeit.Word(), Status: "active"}
	req := model.GetRoomRequest{RoomUUID: roomUUID}

	s.roomRepository.On("GetRoom", context.Background(), req).Return(model.GetRoomResponse{Room: room}, nil)

	res, err := s.service.GetRoom(context.Background(), req)
	s.NoError(err)
	s.Equal(room, res.Room)
}
