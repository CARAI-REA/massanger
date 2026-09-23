package participant

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestRemoveParticipantSelfSuccess() {
	var (
		userUUID  = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()
		room      = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
				AutoClose:       false,
			},
		}
		req = model.RemoveParticipantRequest{
			RoomUUID: roomUUID,
			UserUUID: userUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("RemoveParticipant", ctx, model.RemoveParticipantRequest{
		RoomUUID:  roomUUID,
		UserUUID:  userUUID,
		RemovedBy: userUUID,
	}).Return(model.RemoveParticipantResponse{Success: true}, nil)

	out, err := s.service.RemoveParticipant(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.Success)
}

func (s *ServiceSuite) TestRemoveParticipantKickByOwner() {
	var (
		userUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		room      = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
			},
		}
		req = model.RemoveParticipantRequest{
			RoomUUID:  roomUUID,
			UserUUID:  userUUID,
			RemovedBy: ownerUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("RemoveParticipant", ctx, req).
		Return(model.RemoveParticipantResponse{Success: true}, nil)

	out, err := s.service.RemoveParticipant(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.Success)
}

func (s *ServiceSuite) TestRemoveParticipantKickDenied() {
	var (
		userUUID = gofakeit.UUID()
		actor    = gofakeit.UUID()
		roomUUID = gofakeit.UUID()
		room     = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: gofakeit.UUID(),
			Status:    "active",
		}
		req = model.RemoveParticipantRequest{
			RoomUUID:  roomUUID,
			UserUUID:  userUUID,
			RemovedBy: actor,
		}
	)

	ctx := authctx.WithUserID(context.Background(), actor)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.AssertNotCalled(s.T(), "RemoveParticipant")

	_, err := s.service.RemoveParticipant(ctx, req)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}

func (s *ServiceSuite) TestUpdateParticipantMetadataSuccess() {
	var (
		userUUID = gofakeit.UUID()
		req      = model.UpdateParticipantMetadataRequest{
			RoomUUID: gofakeit.UUID(),
			UserUUID: userUUID,
			Metadata: model.ParticipantMetadata{DisplayName: "n", AudioMuted: true},
		}
		resp = model.UpdateParticipantMetadataResponse{Seccess: true}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.participantRepository.On("UpdateParticipantMetadata", ctx, req).Return(resp, nil)

	out, err := s.service.UpdateParticipantMetadata(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.Seccess)
}

func (s *ServiceSuite) TestUpdateParticipantMetadataDenied() {
	req := model.UpdateParticipantMetadataRequest{
		RoomUUID: gofakeit.UUID(),
		UserUUID: gofakeit.UUID(),
	}
	ctx := authctx.WithUserID(context.Background(), gofakeit.UUID())
	s.participantRepository.AssertNotCalled(s.T(), "UpdateParticipantMetadata")

	_, err := s.service.UpdateParticipantMetadata(ctx, req)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}

func (s *ServiceSuite) TestIsParticipantSuccess() {
	var (
		userUUID  = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()
		room      = model.Room{RoomUUID: roomUUID, OwnerUUID: ownerUUID, Status: "active"}
		req       = model.IsParticipantRequest{RoomUUID: roomUUID, UserUUID: userUUID}
		resp      = model.IsParticipantResponse{IsActive: true}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).
		Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("IsParticipant", ctx, req).Return(resp, nil)

	out, err := s.service.IsParticipant(ctx, req)
	s.Require().NoError(err)
	s.Require().True(out.IsActive)
}
