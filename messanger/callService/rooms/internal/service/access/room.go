package access

import (
	"strings"

	"rooms/internal/model"
)

const roomStatusActive = "active"

// RoomActive комната в статусе active (room_task.md §7.2 п.1).
func RoomActive(room model.Room) bool {
	return strings.EqualFold(room.Status, roomStatusActive)
}

// IsPrivateRoom пустой allowed_users = публичная комната (см. комментарии к model.RoomSettings).
func IsPrivateRoom(settings model.RoomSettings) bool {
	return len(settings.AllowedUsers) > 0
}

// UserMayJoinPrivate A1: приватная комната — только allowed_users; владелец всегда допущен.
func UserMayJoinPrivate(settings model.RoomSettings, room model.Room, userUUID string) bool {
	if !IsPrivateRoom(settings) {
		return true
	}
	if room.OwnerUUID == userUUID {
		return true
	}
	for _, u := range settings.AllowedUsers {
		if u == userUUID {
			return true
		}
	}
	return false
}

// ValidateCreateRoom базовая валидация R1 / InvalidArgument из ТЗ §7.6.
func ValidateCreateRoom(req model.CreateRoomRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return model.ErrInvalidArgument
	}
	if req.RoomSettings.MaxParticipants <= 0 {
		return model.ErrInvalidArgument
	}
	return nil
}
