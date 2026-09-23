package participant

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestGetParticipantSuccessAsOwner() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		userUUID  = gofakeit.UUID()
		joinedAt  = gofakeit.Date()
		displayName = gofakeit.Name()

		participant = model.Participant{
			UserUUID: userUUID,
			RommUUID: roomUUID,
			JoinedAt: lo.ToPtr(joinedAt),
			Metadata: model.ParticipantMetadata{
				DisplayName:   displayName,
				Role:          "participant",
				AudioMuted:    false,
				VideoMuted:    false,
				UserAgent:     gofakeit.UserAgent(),
				ClientVersion: "1.0",
			},
		}

		getParticipantsRequest = model.GetParticipantsRequest{
			RoomUUID:   roomUUID,
			OnlyActive: true,
		}

		getParticipantsResponse = model.GetParticipantsResponse{
			Partisipants: []model.Participant{participant},
			ActiveCount:  1,
		}

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, getParticipantsRequest).Return(getParticipantsResponse, nil)

	res, err := s.service.GetParticipant(ctx, getParticipantsRequest)

	s.Require().NoError(err)
	s.Require().Equal(getParticipantsResponse, res)
}

func (s *ServiceSuite) TestGetParticipantError() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		repoErr   = gofakeit.Error()

		getParticipantsRequest = model.GetParticipantsRequest{
			RoomUUID:   roomUUID,
			OnlyActive: false,
		}

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, getParticipantsRequest).Return(model.GetParticipantsResponse{}, repoErr)

	res, err := s.service.GetParticipant(ctx, getParticipantsRequest)

	s.Require().Error(err)
	s.Require().ErrorIs(err, repoErr)
	s.Require().Empty(res)
}

func (s *ServiceSuite) TestGetParticipantPermissionDenied() {
	var (
		ownerUUID = gofakeit.UUID()
		other     = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()

		getParticipantsRequest = model.GetParticipantsRequest{
			RoomUUID:   roomUUID,
			OnlyActive: true,
		}

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
		}
	)

	ctx := authctx.WithUserID(context.Background(), other)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("IsParticipant", ctx, model.IsParticipantRequest{RoomUUID: roomUUID, UserUUID: other}).Return(model.IsParticipantResponse{IsActive: false}, nil)
	s.participantRepository.AssertNotCalled(s.T(), "GetParticipant")

	_, err := s.service.GetParticipant(ctx, getParticipantsRequest)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}
