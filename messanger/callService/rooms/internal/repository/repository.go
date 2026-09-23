package repository

import (
	"context"
	"time"

	"rooms/internal/model"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, req model.CreateRoomRequest) (model.CreateRoomResponse, error)
	GetRoom(ctx context.Context, req model.GetRoomRequest) (model.GetRoomResponse, error)
	UpdateRoom(ctx context.Context, req model.UpdateRoomRequest) (model.UpdateRoomResponse, error)
	DeleteRoom(ctx context.Context, req model.DeleteRoomRequest) (model.DeleteRoomResponse, error)
	ListRooms(ctx context.Context, req model.ListRoomRequest) (model.ListRoomResponse, error)
	EndRoom(ctx context.Context, req model.EndRoomRequest) (model.EndRoomResponse, error)
}

// RoomCache кэш комнаты и множество активных участников (room_task.md §6.3, §7.3).
type RoomCache interface {
	SetRoom(ctx context.Context, room model.Room, ttl time.Duration) error
	GetRoom(ctx context.Context, roomUUID string) (model.Room, error)
	InvalidateRoomInfo(ctx context.Context, roomUUID string) error
	PurgeRoom(ctx context.Context, roomUUID string) error
	AddActiveParticipant(ctx context.Context, roomUUID, userUUID string) error
	RemoveActiveParticipant(ctx context.Context, roomUUID, userUUID string) error
	ActiveParticipantCount(ctx context.Context, roomUUID string) (int64, error)
}

type ParticipantRepository interface {
	AddParticipant(ctx context.Context, req model.AddParticipantRequest) (model.AddParticipantResponse, error)
	RemoveParticipant(ctx context.Context, req model.RemoveParticipantRequest) (model.RemoveParticipantResponse, error)
	GetParticipant(ctx context.Context, req model.GetParticipantsRequest) (model.GetParticipantsResponse, error)
	IsParticipant(ctx context.Context, req model.IsParticipantRequest) (model.IsParticipantResponse, error)
	UpdateParticipantMetadata(ctx context.Context, req model.UpdateParticipantMetadataRequest) (model.UpdateParticipantMetadataResponse, error)
}
