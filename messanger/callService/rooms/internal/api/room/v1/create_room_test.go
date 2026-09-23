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

func (s *APISuite) TestCreateRoomSuccess() {
	var (
		name      = gofakeit.Word()
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		createdAt = gofakeit.Date()
		updatedAt = gofakeit.Date()
		joinToken = gofakeit.LetterN(32)

		req = &roomsV1.CreateRoomRequest{
			Name:      name,
			OwnerUuid: ownerUUID,
			RoomSettings: &roomsV1.RoomSettings{
				MaxParticipants: 10,
				Quality:         "HD",
				AutoClose:       true,
			},
		}

		expectedReq = converter.CreateRoomRequestToModel(req)

		serviceResp = model.CreateRoomResponse{
			Room: model.Room{
				RoomUUID:     roomUUID,
				Name:         name,
				OwnerUUID:    ownerUUID,
				CratedAt:     lo.ToPtr(createdAt),
				UpdatedAt:    lo.ToPtr(updatedAt),
				RoomSettings: expectedReq.RoomSettings,
				Status:       "active",
			},
			JoinToken: joinToken,
		}
	)

	s.roomService.On("CreateRoom", s.ctx, expectedReq).Return(serviceResp, nil)

	res, err := s.api.CreateRoom(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(joinToken, res.GetJoinToken())
	s.Require().Equal(roomUUID, res.GetRoom().GetRoomUuid())
	s.Require().Equal(name, res.GetRoom().GetName())
}

func (s *APISuite) TestCreateRoomInvalidArgument() {
	var (
		req = &roomsV1.CreateRoomRequest{
			Name:      "",
			OwnerUuid: gofakeit.UUID(),
			RoomSettings: &roomsV1.RoomSettings{
				MaxParticipants: 10,
			},
		}
		expectedReq = converter.CreateRoomRequestToModel(req)
	)

	s.roomService.On("CreateRoom", s.ctx, expectedReq).Return(model.CreateRoomResponse{}, model.ErrInvalidArgument)

	res, err := s.api.CreateRoom(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
}
