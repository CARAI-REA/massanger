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

func (s *APISuite) TestRemoveParticipantSuccess() {
	req := &roomsV1.RemoveParticipantRequest{
		RoomUuid:  gofakeit.UUID(),
		UserUuid:  gofakeit.UUID(),
		RemovedBy: gofakeit.UUID(),
	}
	expected := converter.RemoveParticipantRequestToModel(req)
	s.participantService.On("RemoveParticipant", s.ctx, expected).
		Return(model.RemoveParticipantResponse{Success: true}, nil)

	res, err := s.api.RemoveParticipant(s.ctx, req)
	s.Require().NoError(err)
	s.Require().True(res.GetSuccess())
}

func (s *APISuite) TestGetParticipantsSuccess() {
	roomUUID := gofakeit.UUID()
	req := &roomsV1.GetParticipantsRequest{RoomUuid: roomUUID, OnlyActive: true}
	expected := converter.GetParticipantRequestToModel(req)
	resp := model.GetParticipantsResponse{
		Partisipants: []model.Participant{{
			UserUUID: gofakeit.UUID(),
			RommUUID: roomUUID,
			JoinedAt: lo.ToPtr(gofakeit.Date()),
			Metadata: model.ParticipantMetadata{DisplayName: "u"},
		}},
		ActiveCount: 1,
	}
	s.participantService.On("GetParticipant", s.ctx, expected).Return(resp, nil)

	res, err := s.api.GetParticipants(s.ctx, req)
	s.Require().NoError(err)
	s.Require().Equal(int32(1), res.GetActiveCount())
	s.Require().Len(res.GetParticipants(), 1)
}

func (s *APISuite) TestIsParticipantSuccess() {
	req := &roomsV1.IsParticipantRequest{
		RoomUuid: gofakeit.UUID(),
		UserUuid: gofakeit.UUID(),
	}
	expected := converter.IsParticipantRequestToModel(req)
	resp := model.IsParticipantResponse{
		IsActive: true,
		Participant: model.Participant{
			UserUUID: req.UserUuid,
			RommUUID: req.RoomUuid,
			JoinedAt: lo.ToPtr(gofakeit.Date()),
		},
	}
	s.participantService.On("IsParticipant", s.ctx, expected).Return(resp, nil)

	res, err := s.api.IsParticipant(s.ctx, req)
	s.Require().NoError(err)
	s.Require().True(res.GetIsActive())
}

func (s *APISuite) TestUpdateParticipantMetadataSuccess() {
	req := &roomsV1.UpdateParticipantMetadataRequest{
		RoomUuid: gofakeit.UUID(),
		UserUuid: gofakeit.UUID(),
		Metadata: &roomsV1.ParticipantMetadata{
			DisplayName: "new",
			AudioMuted:  true,
		},
	}
	expected := converter.UpdateParticipantMetadataRequestToModel(req)
	resp := model.UpdateParticipantMetadataResponse{
		Seccess: true,
		Participant: model.Participant{
			UserUUID: req.UserUuid,
			RommUUID: req.RoomUuid,
			JoinedAt: lo.ToPtr(gofakeit.Date()),
			Metadata: expected.Metadata,
		},
	}
	s.participantService.On("UpdateParticipantMetadata", s.ctx, expected).Return(resp, nil)

	res, err := s.api.UpdateParticipantMetadata(s.ctx, req)
	s.Require().NoError(err)
	s.Require().True(res.GetSuccess())
	s.Require().Equal("new", res.GetParticipant().GetMetadata().GetDisplayName())
}

func (s *APISuite) TestUpdateParticipantMetadataConflict() {
	req := &roomsV1.UpdateParticipantMetadataRequest{
		RoomUuid: gofakeit.UUID(),
		UserUuid: gofakeit.UUID(),
		Metadata: &roomsV1.ParticipantMetadata{DisplayName: "x"},
	}
	expected := converter.UpdateParticipantMetadataRequestToModel(req)
	s.participantService.On("UpdateParticipantMetadata", s.ctx, expected).
		Return(model.UpdateParticipantMetadataResponse{}, model.ErrParticipantNotFound)

	res, err := s.api.UpdateParticipantMetadata(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.NotFound, st.Code())
}
