package participant

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"

	"rooms/internal/model"
	"rooms/internal/service/authctx"
)

func (s *ServiceSuite) TestAddParticipantSuccess() {
	var (
		userUUID  = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
				AllowedUsers:    nil,
				Quality:         "hd",
			},
		}

		addReq = model.AddParticipantRequest{
			RoomUUID: roomUUID,
			UserUUID: userUUID,
			Metadata: model.ParticipantMetadata{DisplayName: "u"},
		}

		addResp = model.AddParticipantResponse{Success: true, JoinToken: "jwt"}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, model.GetParticipantsRequest{RoomUUID: roomUUID, OnlyActive: true}).Return(model.GetParticipantsResponse{ActiveCount: 3}, nil)
	s.participantRepository.On("IsParticipant", ctx, model.IsParticipantRequest{RoomUUID: roomUUID, UserUUID: userUUID}).Return(model.IsParticipantResponse{IsActive: false}, nil)
	s.participantRepository.On("AddParticipant", ctx, addReq).Return(addResp, nil)

	res, err := s.service.AddParticipant(ctx, addReq)
	s.Require().NoError(err)
	s.Require().Equal(addResp, res)
}

func (s *ServiceSuite) TestAddParticipantResourceExhausted() {
	var (
		userUUID  = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 2,
			},
		}

		addReq = model.AddParticipantRequest{
			RoomUUID: roomUUID,
			UserUUID: userUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, model.GetParticipantsRequest{RoomUUID: roomUUID, OnlyActive: true}).Return(model.GetParticipantsResponse{ActiveCount: 2}, nil)
	s.participantRepository.AssertNotCalled(s.T(), "AddParticipant")

	_, err := s.service.AddParticipant(ctx, addReq)
	s.Require().ErrorIs(err, model.ErrResourceExhausted)
}

func (s *ServiceSuite) TestAddParticipantPrivateDenied() {
	var (
		userUUID  = gofakeit.UUID()
		other     = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
				AllowedUsers:    []string{other},
			},
		}

		addReq = model.AddParticipantRequest{
			RoomUUID: roomUUID,
			UserUUID: userUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, model.GetParticipantsRequest{RoomUUID: roomUUID, OnlyActive: true}).Return(model.GetParticipantsResponse{ActiveCount: 1}, nil)
	s.participantRepository.AssertNotCalled(s.T(), "AddParticipant")

	_, err := s.service.AddParticipant(ctx, addReq)
	s.Require().ErrorIs(err, model.ErrPermissionDenied)
}

func (s *ServiceSuite) TestAddParticipantOwnerAllowedInPrivate() {
	var (
		ownerUUID = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
				AllowedUsers:    []string{gofakeit.UUID()},
			},
		}

		addReq = model.AddParticipantRequest{
			RoomUUID: roomUUID,
			UserUUID: ownerUUID,
		}

		addResp = model.AddParticipantResponse{Success: true}
	)

	ctx := authctx.WithUserID(context.Background(), ownerUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, model.GetParticipantsRequest{RoomUUID: roomUUID, OnlyActive: true}).Return(model.GetParticipantsResponse{ActiveCount: 0}, nil)
	s.participantRepository.On("IsParticipant", ctx, model.IsParticipantRequest{RoomUUID: roomUUID, UserUUID: ownerUUID}).Return(model.IsParticipantResponse{IsActive: false}, nil)
	s.participantRepository.On("AddParticipant", ctx, addReq).Return(addResp, nil)

	res, err := s.service.AddParticipant(ctx, addReq)
	s.Require().NoError(err)
	s.Require().Equal(addResp, res)
}

func (s *ServiceSuite) TestAddParticipantAlreadyActive() {
	var (
		userUUID  = gofakeit.UUID()
		roomUUID  = gofakeit.UUID()
		ownerUUID = gofakeit.UUID()

		room = model.Room{
			RoomUUID:  roomUUID,
			OwnerUUID: ownerUUID,
			Status:    "active",
			RoomSettings: model.RoomSettings{
				MaxParticipants: 10,
			},
		}

		addReq = model.AddParticipantRequest{
			RoomUUID: roomUUID,
			UserUUID: userUUID,
		}
	)

	ctx := authctx.WithUserID(context.Background(), userUUID)
	s.roomRepository.On("GetRoom", ctx, model.GetRoomRequest{RoomUUID: roomUUID}).Return(model.GetRoomResponse{Room: room}, nil)
	s.participantRepository.On("GetParticipant", ctx, model.GetParticipantsRequest{RoomUUID: roomUUID, OnlyActive: true}).Return(model.GetParticipantsResponse{ActiveCount: 1}, nil)
	s.participantRepository.On("IsParticipant", ctx, model.IsParticipantRequest{RoomUUID: roomUUID, UserUUID: userUUID}).Return(model.IsParticipantResponse{IsActive: true}, nil)
	s.participantRepository.AssertNotCalled(s.T(), "AddParticipant")

	_, err := s.service.AddParticipant(ctx, addReq)
	s.Require().ErrorIs(err, model.ErrParticipantConflict)
}
