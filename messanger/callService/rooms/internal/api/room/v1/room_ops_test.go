package v1

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"

	"rooms/internal/converter"
	"rooms/internal/model"
)

func (s *APISuite) TestUpdateRoomSuccess() {
	var (
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()
		req       = &roomsV1.UpdateRoomRequest{
			RoomUuid:  roomUUID,
			OwnerUuid: ownerUUID,
			RoomSettings: &roomsV1.RoomSettings{
				MaxParticipants: 20,
				Quality:         "SD",
			},
		}
		expected = converter.UpdateRoomRequestToModel(req)
		resp     = model.UpdateRoomResponse{
			Success: true,
			Room: model.Room{
				RoomUUID:     roomUUID,
				OwnerUUID:    ownerUUID,
				CratedAt:     lo.ToPtr(gofakeit.Date()),
				UpdatedAt:    lo.ToPtr(gofakeit.Date()),
				RoomSettings: expected.RoomSettings,
				Status:       "active",
			},
		}
	)

	s.roomService.On("UpdateRoom", s.ctx, expected).Return(resp, nil)

	res, err := s.api.UpdateRoom(s.ctx, req)
	s.Require().NoError(err)
	s.Require().True(res.GetSuccess())
	s.Require().Equal(roomUUID, res.GetRoom().GetRoomUuid())
}

func (s *APISuite) TestDeleteRoomSuccess() {
	req := &roomsV1.DeleteRoomRequest{
		RoomUuid:  gofakeit.UUID(),
		OwnerUuid: gofakeit.UUID(),
		Permanent: false,
	}
	expected := converter.DeleteRoomRequestToModel(req)
	s.roomService.On("DeleteRoom", s.ctx, expected).Return(model.DeleteRoomResponse{Success: true}, nil)

	res, err := s.api.DeleteRoom(s.ctx, req)
	s.Require().NoError(err)
	s.Require().True(res.GetSuccess())
}

func (s *APISuite) TestDeleteRoomPermissionDenied() {
	req := &roomsV1.DeleteRoomRequest{
		RoomUuid:  gofakeit.UUID(),
		OwnerUuid: gofakeit.UUID(),
	}
	expected := converter.DeleteRoomRequestToModel(req)
	s.roomService.On("DeleteRoom", s.ctx, expected).Return(model.DeleteRoomResponse{}, model.ErrPermissionDenied)

	res, err := s.api.DeleteRoom(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.PermissionDenied, st.Code())
}

func (s *APISuite) TestListRoomsSuccess() {
	ownerUUID := gofakeit.UUID()
	req := &roomsV1.ListRoomsRequest{
		OwnerUuid: ownerUUID,
		Status:    "active",
		Limit:     10,
		Offset:    0,
	}
	expected := converter.ListRoomRequestToModel(req)
	resp := model.ListRoomResponse{
		Rooms: []model.Room{{
			RoomUUID:  gofakeit.UUID(),
			OwnerUUID: ownerUUID,
			Name:      gofakeit.Word(),
			CratedAt:  lo.ToPtr(gofakeit.Date()),
			UpdatedAt: lo.ToPtr(gofakeit.Date()),
			Status:    "active",
		}},
		Total: 1,
	}
	s.roomService.On("ListRooms", s.ctx, expected).Return(resp, nil)

	res, err := s.api.ListRooms(s.ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(int32(1), res.GetTotal())
	s.Require().Len(res.GetRooms(), 1)
}

func (s *APISuite) TestEndRoomSuccess() {
	req := &roomsV1.EndRoomRequest{
		RoomUuid:  gofakeit.UUID(),
		OwnerUuid: gofakeit.UUID(),
	}
	expected := converter.EndRoomRequestToModel(req)
	s.roomService.On("EndRoom", s.ctx, expected).Return(model.EndRoomResponse{Success: true}, nil)

	res, err := s.api.EndRoom(s.ctx, req)
	s.Require().NoError(err)
	s.Require().True(res.GetSuccess())
}
