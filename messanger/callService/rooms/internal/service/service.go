package service

import (
	"context"

	"rooms/internal/model"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req model.CreateRoomRequest) (model.CreateRoomResponse, error)
	GetRoom(ctx context.Context, req model.GetRoomRequest) (model.GetRoomResponse, error)
	UpdateRoom(ctx context.Context, req model.UpdateRoomRequest) (model.UpdateRoomResponse, error)
	DeleteRoom(ctx context.Context, req model.DeleteRoomRequest) (model.DeleteRoomResponse, error)
	ListRooms(ctx context.Context, req model.ListRoomRequest) (model.ListRoomResponse, error)
	EndRoom(ctx context.Context, req model.EndRoomRequest) (model.EndRoomResponse, error)
}

type ParticipantService interface {
	AddParticipant(ctx context.Context, req model.AddParticipantRequest) (model.AddParticipantResponse, error)
	RemoveParticipant(ctx context.Context, req model.RemoveParticipantRequest) (model.RemoveParticipantResponse, error)
	GetParticipant(ctx context.Context, req model.GetParticipantsRequest) (model.GetParticipantsResponse, error)
	IsParticipant(ctx context.Context, req model.IsParticipantRequest) (model.IsParticipantResponse, error)
	UpdateParticipantMetadata(ctx context.Context, req model.UpdateParticipantMetadataRequest) (model.UpdateParticipantMetadataResponse, error)
	AssertCanJoin(ctx context.Context, req model.AssertCanJoinRequest) (model.AssertCanJoinResponse, error)
	RefreshJoinToken(ctx context.Context, req model.RefreshJoinTokenRequest) (model.RefreshJoinTokenResponse, error)
	GetTURNCredentials(ctx context.Context, req model.GetTURNCredentialsRequest) (model.GetTURNCredentialsResponse, error)
}
