package room

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestUpdateRoomSuccess() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		room      = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
		}
		req = model.UpdateRoomRequest{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			RoomSettings: model.RoomSettings{
				MaxParticipants: 12,
				Quality:         "hd",
			},
		}
		resp = model.UpdateRoomResponse{Success: true, Room: room}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.roomRepository.On("UpdateRoom", ctx, req).Return(resp, nil)

	out, err := s.service.UpdateRoom(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.Success)
}

func (s *ServiceSuite) TestDeleteRoomSuccess() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		room      = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
		}
		req = model.DeleteRoomRequest{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.roomRepository.On("DeleteRoom", ctx, req).Return(model.DeleteRoomResponse{Success: true}, nil)

	out, err := s.service.DeleteRoom(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.Success)
}

func (s *ServiceSuite) TestDeleteRoomNotOwner() {
	var (
		ownerUUID = gofakeit.UUID()
		actor     = gofakeit.UUID()
		req       = model.DeleteRoomRequest{
			RoomUUID:  gofakeit.UUID(),
			OwnerUUID: ownerUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), actor)
	s.roomRepository.AssertNotCalled(s.T(), "DeleteRoom")

	_, err := s.service.DeleteRoom(ctx, req)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}

func (s *ServiceSuite) TestEndRoomSuccess() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		room      = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
		}
		req = model.EndRoomRequest{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.roomRepository.On("EndRoom", ctx, req).Return(model.EndRoomResponse{Success: true}, nil)

	out, err := s.service.EndRoom(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.Success)
}

func (s *ServiceSuite) TestEndRoomInactive() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		room      = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "ended",
		}
		req = model.EndRoomRequest{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.roomRepository.AssertNotCalled(s.T(), "EndRoom")

	_, err := s.service.EndRoom(ctx, req)
	s.Require().ErrorIs(err, model.ErrRoomInactive)
}
