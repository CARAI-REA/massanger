package participant

import (
	"rooms/internal/model"
	"rooms/internal/service/authctx"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/mock"
)

func (s *ServiceSuite) TestAssertCanJoinActiveParticipant() {
	roomUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()

	s.roomRepository.On("GetRoom", mock.Anything, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: model.Room{RoomUUID: roomUUID, Status: "active"}}, nil)
	s.participantRepository.On("IsParticipant", mock.Anything, model.IsParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	}).Return(model.IsParticipantResponse{IsActive: true}, nil)

	resp, err := s.service.AssertCanJoin(s.ctx, model.AssertCanJoinRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.Require().NoError(err)
	s.True(resp.Ok)
	s.Equal("active", resp.RoomStatus)
}

func (s *ServiceSuite) TestAssertCanJoinInactiveRoom() {
	roomUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()

	s.roomRepository.On("GetRoom", mock.Anything, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: model.Room{RoomUUID: roomUUID, Status: "ended"}}, nil)
	s.participantRepository.On("IsParticipant", mock.Anything, model.IsParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	}).Return(model.IsParticipantResponse{IsActive: true}, nil)

	resp, err := s.service.AssertCanJoin(s.ctx, model.AssertCanJoinRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.Require().NoError(err)
	s.False(resp.Ok)
	s.Equal("ended", resp.RoomStatus)
}

func (s *ServiceSuite) TestAssertCanJoinNotParticipant() {
	roomUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()

	s.roomRepository.On("GetRoom", mock.Anything, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: model.Room{RoomUUID: roomUUID, Status: "active"}}, nil)
	s.participantRepository.On("IsParticipant", mock.Anything, model.IsParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	}).Return(model.IsParticipantResponse{IsActive: false}, nil)

	resp, err := s.service.AssertCanJoin(s.ctx, model.AssertCanJoinRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.Require().NoError(err)
	s.False(resp.Ok)
}

func (s *ServiceSuite) TestAssertCanJoinRoomNotFound() {
	roomUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()

	s.roomRepository.On("GetRoom", mock.Anything, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{}, model.ErrRoomNotFound)

	_, err := s.service.AssertCanJoin(s.ctx, model.AssertCanJoinRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.ErrorIs(err, model.ErrRoomNotFound)
}

func (s *ServiceSuite) TestRefreshJoinTokenSuccess() {
	roomUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()
	ctx := authctx.WithUserID(s.ctx, userUUID)

	s.participantRepository.On("IsParticipant", mock.Anything, model.IsParticipantRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	}).Return(model.IsParticipantResponse{IsActive: true}, nil)

	resp, err := s.service.RefreshJoinToken(ctx, model.RefreshJoinTokenRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.Require().NoError(err)
	s.Equal("join-"+userUUID+"-"+roomUUID, resp.JoinToken)
}

func (s *ServiceSuite) TestRefreshJoinTokenWrongUser() {
	roomUUID := gofakeit.UUID()
	userUUID := gofakeit.UUID()
	ctx := authctx.WithUserID(s.ctx, gofakeit.UUID())

	_, err := s.service.RefreshJoinToken(ctx, model.RefreshJoinTokenRequest{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
	})
	s.ErrorIs(err, model.ErrPermissionDenied)
}
