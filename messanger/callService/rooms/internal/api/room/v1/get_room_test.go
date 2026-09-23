package v1

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rooms/internal/converter"
	"rooms/internal/model"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func (s *APISuite) TestGetRoomSuccess() {
	var (
		roomUUID  = gofakeit.UUID()
		name      = gofakeit.Word()
		ownerUUID = gofakeit.UUID()
		createdAt = gofakeit.Date()
		updatedAt = gofakeit.Date()

		req = &roomsV1.GetRoomRequest{
			RoomUuid: roomUUID,
		}

		expectedReq = converter.GetRoomRequestToModel(req)

		serviceResp = model.GetRoomResponse{
			Room: model.Room{
				RoomUUID:  roomUUID,
				Name:      name,
				OwnerUUID: ownerUUID,
				CratedAt:  lo.ToPtr(createdAt),
				UpdatedAt: lo.ToPtr(updatedAt),
				RoomSettings: model.RoomSettings{
					MaxParticipants: 10,
					Quality:         "HD",
				},
				Status: "active",
			},
		}
	)

	s.roomService.On("GetRoom", s.ctx, expectedReq).Return(serviceResp, nil)

	res, err := s.api.GetRoom(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(roomUUID, res.GetRoom().GetRoomUuid())
	s.Require().Equal(name, res.GetRoom().GetName())
}

func (s *APISuite) TestGetRoomNotFound() {
	var (
		roomUUID = gofakeit.UUID()
		req      = &roomsV1.GetRoomRequest{RoomUuid: roomUUID}
		expected = converter.GetRoomRequestToModel(req)
	)

	s.roomService.On("GetRoom", s.ctx, expected).Return(model.GetRoomResponse{}, model.ErrRoomNotFound)

	res, err := s.api.GetRoom(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.NotFound, st.Code())
}

func (s *APISuite) TestGetRoomError() {
	var (
		serviceErr = gofakeit.Error()
		roomUUID   = gofakeit.UUID()
		req        = &roomsV1.GetRoomRequest{RoomUuid: roomUUID}
		expected   = converter.GetRoomRequestToModel(req)
	)

	s.roomService.On("GetRoom", s.ctx, expected).Return(model.GetRoomResponse{}, serviceErr)

	res, err := s.api.GetRoom(s.ctx, req)
	s.Require().Error(err)
	s.Require().ErrorIs(err, serviceErr)
	s.Require().Nil(res)
}
