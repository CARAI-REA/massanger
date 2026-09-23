package room

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestListRoomsSuccess() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		name      = gofakeit.Word()
		status    = "active"
		createdAt = gofakeit.Date()
		updatedAt = gofakeit.Date()

		room = model.Room{
			RoomUUID:  roomUUID,
			Name:      name,
			OwnerUUID: ownerUUID,
			CratedAt:  lo.ToPtr(createdAt),
			UpdatedAt: lo.ToPtr(updatedAt),
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
				Quality:         "hd",
				AutoClose:       false,
			},
			Status: status,
		}

		listRoomsResponse = model.ListRoomResponse{
			Rooms: []model.Room{room},
			Total: 1,
		}

		listRoomsRequest = model.ListRoomRequest{
			OwnerUUID: ownerUUID,
			Status:    status,
			Limit:     20,
			Offset:    0,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("ListRooms", ctx, listRoomsRequest).Return(listRoomsResponse, nil)

	res, err := s.service.ListRooms(ctx, listRoomsRequest)

	s.Require().NoError(err)
	s.Require().Equal(listRoomsResponse, res)
}

func (s *ServiceSuite) TestListRoomsError() {
	var (
		ownerUUID = gofakeit.UUID()
		repoErr   = gofakeit.Error()

		listRoomsRequest = model.ListRoomRequest{
			OwnerUUID: ownerUUID,
			Status:    "all",
			Limit:     10,
			Offset:    0,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("ListRooms", ctx, listRoomsRequest).Return(model.ListRoomResponse{}, repoErr)

	res, err := s.service.ListRooms(ctx, listRoomsRequest)

	s.Require().Error(err)
	s.Require().ErrorIs(err, repoErr)
	s.Require().Empty(res)
}

func (s *ServiceSuite) TestListRoomsEmpty() {
	var (
		ownerUUID = gofakeit.UUID()

		listRoomsRequest = model.ListRoomRequest{
			OwnerUUID: ownerUUID,
			Status:    "active",
			Limit:     10,
			Offset:    0,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("ListRooms", ctx, listRoomsRequest).Return(model.ListRoomResponse{}, nil)

	res, err := s.service.ListRooms(ctx, listRoomsRequest)

	s.Require().NoError(err)
	s.Require().Empty(res.Rooms)
	s.Require().Equal(model.ListRoomResponse{}, res)
}

func (s *ServiceSuite) TestListRoomsPermissionDenied() {
	var (
		ownerUUID = gofakeit.UUID()
		other     = gofakeit.UUID()

		listRoomsRequest = model.ListRoomRequest{
			OwnerUUID: ownerUUID,
			Status:    "active",
			Limit:     10,
			Offset:    0,
		}
	)

	ctx := authctx.WithUserID(context.Background(), other)
	s.roomRepository.AssertNotCalled(s.T(), "ListRooms")

	_, err := s.service.ListRooms(ctx, listRoomsRequest)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}
