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

func (s *APISuite) TestAddParticipantSuccess() {
	var (
		roomUUID  = gofakeit.UUID()
		userUUID  = gofakeit.UUID()
		joinToken = gofakeit.LetterN(32)
		joinedAt  = gofakeit.Date()

		req = &roomsV1.AddParticipantRequest{
			RoomUuid: roomUUID,
			UserUuid: userUUID,
			Metadata: &roomsV1.ParticipantMetadata{
				DisplayName: gofakeit.Name(),
				Role:        "participant",
			},
		}

		expectedReq = converter.AddParticipantRequestToModel(req)

		serviceResp = model.AddParticipantResponse{
			Success:   true,
			JoinToken: joinToken,
			Participant: model.Participant{
				UserUUID: userUUID,
				RommUUID: roomUUID,
				JoinedAt: lo.ToPtr(joinedAt),
				Metadata: expectedReq.Metadata,
			},
		}
	)

	s.participantService.On("AddParticipant", s.ctx, expectedReq).Return(serviceResp, nil)

	res, err := s.api.AddParticipant(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().True(res.GetSuccess())
	s.Require().Equal(joinToken, res.GetJoinToken())
	s.Require().Equal(userUUID, res.GetParticipant().GetUserUuid())
}

func (s *APISuite) TestAddParticipantResourceExhausted() {
	var (
		req = &roomsV1.AddParticipantRequest{
			RoomUuid: gofakeit.UUID(),
			UserUuid: gofakeit.UUID(),
		}
		expected = converter.AddParticipantRequestToModel(req)
	)

	s.participantService.On("AddParticipant", s.ctx, expected).
		Return(model.AddParticipantResponse{}, model.ErrResourceExhausted)

	res, err := s.api.AddParticipant(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(res)

	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.ResourceExhausted, st.Code())
}
