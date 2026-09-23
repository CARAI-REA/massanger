package room

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestCreateRoomSuccess() {
	var (
		ownerUUID = gofakeit.UUID()
		req       = model.CreateRoomRequest{
			Name:      gofakeit.Word(),
			OwnerUUID: ownerUUID,
			RoomSettings: model.RoomSettings{
				MaxParticipants: 8,
				Quality:         "hd",
				AutoClose:       false,
			},
		}
		resp = model.CreateRoomResponse{
			Room: model.Room{
				RoomUUID:  gofakeit.UUID(),
				Name:      req.Name,
				OwnerUUID: ownerUUID,
				Status:    "active",
			},
			JoinToken: "token",
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("CreateRoom", ctx, req).Return(resp, nil)

	out, err := s.service.CreateRoom(ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(resp, out)
}

func (s *ServiceSuite) TestCreateRoomOwnerMismatch() {
	var (
		ownerUUID = gofakeit.UUID()
		actor     = gofakeit.UUID()
		req       = model.CreateRoomRequest{
			Name:      gofakeit.Word(),
			OwnerUUID: ownerUUID,
			RoomSettings: model.RoomSettings{
				MaxParticipants: 5,
				Quality:         "sd",
			},
		}
	)

	ctx := authctx.WithUserID(context.Background(), actor)
	s.roomRepository.AssertNotCalled(s.T(), "CreateRoom")

	_, err := s.service.CreateRoom(ctx, req)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}

func (s *ServiceSuite) TestCreateRoomInvalidName() {
	ownerUUID := gofakeit.UUID()
	req := model.CreateRoomRequest{
		Name:      "   ",
		OwnerUUID: ownerUUID,
		RoomSettings: model.RoomSettings{
			MaxParticipants: 5,
		},
	}
	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.AssertNotCalled(s.T(), "CreateRoom")

	_, err := s.service.CreateRoom(ctx, req)
	s.Require().ErrorIs(err, model.ErrInvalidArgument)
}

func (s *ServiceSuite) TestCreateRoomInvalidMaxParticipants() {
	ownerUUID := gofakeit.UUID()
	req := model.CreateRoomRequest{
		Name:      gofakeit.Word(),
		OwnerUUID: ownerUUID,
		RoomSettings: model.RoomSettings{
			MaxParticipants: 0,
		},
	}
	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.AssertNotCalled(s.T(), "CreateRoom")

	_, err := s.service.CreateRoom(ctx, req)
	s.Require().ErrorIs(err, model.ErrInvalidArgument)
}
