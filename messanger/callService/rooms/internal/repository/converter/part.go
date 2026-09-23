package converter

import (
	model "rooms/internal/model"
	repoModel "rooms/internal/repository/model"
)

// RoomToRepoModel конвертирует doman-модель комнаты в модель репозитория.
func RoomToRepoModel(info model.Room) repoModel.Room {
	return repoModel.Room{
		RoomUUID:     info.RoomUUID,
		Name:         info.Name,
		OwnerUUID:    info.OwnerUUID,
		CreatedAt:    info.CratedAt,
		UpdatedAt:    info.UpdatedAt,
		DeletedAt:    info.DeletedAt,
		RoomSettings: RoomSettingsToRepoModel(info.RoomSettings),
		Status:       info.Status,
	}
}

// RoomSettingsToRepoModel конвертирует doman-настройки комнаты в модель репозитория.
func RoomSettingsToRepoModel(info model.RoomSettings) repoModel.RoomSettings {
	return repoModel.RoomSettings{
		MaxParticipants:  info.MaxParticipants,
		AllowedUsers:     info.AllowedUsers,
		Quality:          info.Quality,
		AutoClose:        info.AutoClose,
		RecordingEnabled: info.RecordingEnabled,
	}
}

// RoomSettingsToModel конвертирует настройки комнаты из модели репозитория в доменную модель.
func RoomSettingsToModel(info repoModel.RoomSettings) model.RoomSettings {
	return model.RoomSettings{
		MaxParticipants:  info.MaxParticipants,
		AllowedUsers:     info.AllowedUsers,
		Quality:          info.Quality,
		AutoClose:        info.AutoClose,
		RecordingEnabled: info.RecordingEnabled,
	}
}

// RoomToModel конвертирует комнату из модели репозитория в доменную модель.
func RoomToModel(info repoModel.Room) model.Room {
	return model.Room{
		RoomUUID:     info.RoomUUID,
		Name:         info.Name,
		OwnerUUID:    info.OwnerUUID,
		CratedAt:     info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
		DeletedAt:    info.DeletedAt,
		RoomSettings: RoomSettingsToModel(info.RoomSettings),
		Status:       info.Status,
	}
}

// CreateRoomRequestToModel конвертирует запрос создания комнаты из репозитория в доменную модель.
func CreateRoomRequestToModel(info repoModel.CreateRoomRequest) model.CreateRoomRequest {
	return model.CreateRoomRequest{
		Name:         info.Name,
		OwnerUUID:    info.OwnerUUID,
		RoomSettings: RoomSettingsToModel(info.RoomSettings),
	}
}

// CreateRoomRequestToRepoModel конвертирует доменную модель запроса создания комнаты в модель репозитория.
func CreateRoomRequestToRepoModel(info model.CreateRoomRequest) repoModel.CreateRoomRequest {
	return repoModel.CreateRoomRequest{
		Name:         info.Name,
		OwnerUUID:    info.OwnerUUID,
		RoomSettings: RoomSettingsToRepoModel(info.RoomSettings),
	}
}

// CreateRoomResponseToModel конвертирует ответ создания комнаты из репозитория в доменную модель.
func CreateRoomResponseToModel(info repoModel.CreateRoomResponse) model.CreateRoomResponse {
	return model.CreateRoomResponse{
		Room:      RoomToModel(info.Room),
		JoinToken: info.JoinToken,
	}
}

// CreateRoomResponseToRepoModel конвертирует доменную модель ответа создания комнаты в модель репозитория.
func CreateRoomResponseToRepoModel(info model.CreateRoomResponse) repoModel.CreateRoomResponse {
	return repoModel.CreateRoomResponse{
		Room:      RoomToRepoModel(info.Room),
		JoinToken: info.JoinToken,
	}
}

// GetRoomRequestToModel конвертирует запрос получения комнаты из репозитория в доменную модель.
func GetRoomRequestToModel(info repoModel.GetRoomRequest) model.GetRoomRequest {
	return model.GetRoomRequest{
		RoomUUID: info.RoomUUID,
	}
}

// GetRoomRequestToRepoModel конвертирует доменную модель запроса получения комнаты в модель репозитория.
func GetRoomRequestToRepoModel(info model.GetRoomRequest) repoModel.GetRoomRequest {
	return repoModel.GetRoomRequest{
		RoomUUID: info.RoomUUID,
	}
}

// GetRoomResponseToModel конвертирует ответ получения комнаты из репозитория в доменную модель.
func GetRoomResponseToModel(info repoModel.GetRoomResponse) model.GetRoomResponse {
	return model.GetRoomResponse{
		Room: RoomToModel(info.Room),
	}
}

// GetRoomResponseToRepoModel конвертирует доменную модель ответа получения комнаты в модель репозитория.
func GetRoomResponseToRepoModel(info model.GetRoomResponse) repoModel.GetRoomResponse {
	return repoModel.GetRoomResponse{
		Room: RoomToRepoModel(info.Room),
	}
}

// UpdateRoomRequestToModel конвертирует запрос обновления комнаты из репозитория в доменную модель.
func UpdateRoomRequestToModel(info repoModel.UpdateRoomRequest) model.UpdateRoomRequest {
	return model.UpdateRoomRequest{
		RoomUUID:     info.RoomUUID,
		OwnerUUID:    info.OwnerUUID,
		RoomSettings: RoomSettingsToModel(info.RoomSettings),
	}
}

// UpdateRoomRequestToRepoModel конвертирует доменную модель запроса обновления комнаты в модель репозитория.
func UpdateRoomRequestToRepoModel(info model.UpdateRoomRequest) repoModel.UpdateRoomRequest {
	return repoModel.UpdateRoomRequest{
		RoomUUID:     info.RoomUUID,
		OwnerUUID:    info.OwnerUUID,
		RoomSettings: RoomSettingsToRepoModel(info.RoomSettings),
	}
}

// UpdateRoomResponseToModel конвертирует ответ обновления комнаты из репозитория в доменную модель.
func UpdateRoomResponseToModel(info repoModel.UpdateRoomResponse) model.UpdateRoomResponse {
	return model.UpdateRoomResponse{
		Success: info.Success,
		Room:    RoomToModel(info.Room),
	}
}

// UpdateRoomResponseToRepoModel конвертирует доменную модель ответа обновления комнаты в модель репозитория.
func UpdateRoomResponseToRepoModel(info model.UpdateRoomResponse) repoModel.UpdateRoomResponse {
	return repoModel.UpdateRoomResponse{
		Success: info.Success,
		Room:    RoomToRepoModel(info.Room),
	}
}

// DeleteRoomRequestToModel конвертирует запрос удаления комнаты из репозитория в доменную модель.
func DeleteRoomRequestToModel(info repoModel.DeleteRoomRequest) model.DeleteRoomRequest {
	return model.DeleteRoomRequest{
		RoomUUID:  info.RoomUUID,
		OwnerUUID: info.OwnerUUID,
		Permanent: info.Permanent,
	}
}

// DeleteRoomRequestToRepoModel конвертирует доменную модель запроса удаления комнаты в модель репозитория.
func DeleteRoomRequestToRepoModel(info model.DeleteRoomRequest) repoModel.DeleteRoomRequest {
	return repoModel.DeleteRoomRequest{
		RoomUUID:  info.RoomUUID,
		OwnerUUID: info.OwnerUUID,
		Permanent: info.Permanent,
	}
}

// DeleteRoomResponseToModel конвертирует ответ удаления комнаты из репозитория в доменную модель.
func DeleteRoomResponseToModel(info repoModel.DeleteRoomResponse) model.DeleteRoomResponse {
	return model.DeleteRoomResponse{
		Success: info.Success,
	}
}

// DeleteRoomResponseToRepoModel конвертирует доменную модель ответа удаления комнаты в модель репозитория.
func DeleteRoomResponseToRepoModel(info model.DeleteRoomResponse) repoModel.DeleteRoomResponse {
	return repoModel.DeleteRoomResponse{
		Success: info.Success,
	}
}

// ListRoomRequestToModel конвертирует запрос списка комнат из репозитория в доменную модель.
func ListRoomRequestToModel(info repoModel.ListRoomRequest) model.ListRoomRequest {
	return model.ListRoomRequest{
		OwnerUUID: info.OwnerUUID,
		Status:    info.Status,
		Limit:     info.Limit,
		Offset:    info.Offset,
	}
}

// ListRoomRequestToRepoModel конвертирует доменную модель запроса списка комнат в модель репозитория.
func ListRoomRequestToRepoModel(info model.ListRoomRequest) repoModel.ListRoomRequest {
	return repoModel.ListRoomRequest{
		OwnerUUID: info.OwnerUUID,
		Status:    info.Status,
		Limit:     info.Limit,
		Offset:    info.Offset,
	}
}

// RoomsSliceToModel конвертирует слайс комнат из модели репозитория в доменную модель.
func RoomsSliceToModel(info []repoModel.Room) []model.Room {
	result := make([]model.Room, len(info))

	for i, room := range info {
		result[i] = RoomToModel(room)
	}
	return result
}

// RoomsSliceToRepoModel конвертирует слайс доменных комнат в слайс моделей репозитория.
func RoomsSliceToRepoModel(info []model.Room) []repoModel.Room {
	result := make([]repoModel.Room, len(info))

	for i, room := range info {
		result[i] = RoomToRepoModel(room)
	}

	return result
}

// ListRoomsResponseToModel конвертирует ответ списка комнат из репозитория в доменную модель.
func ListRoomsResponseToModel(info repoModel.ListRoomResponse) model.ListRoomResponse {
	return model.ListRoomResponse{
		Rooms: RoomsSliceToModel(info.Rooms),
		Total: info.Total,
	}
}

// ListRoomsResponseToRepoModel конвертирует доменную модель ответа списка комнат в модель репозитория.
func ListRoomsResponseToRepoModel(info model.ListRoomResponse) repoModel.ListRoomResponse {
	return repoModel.ListRoomResponse{
		Rooms: RoomsSliceToRepoModel(info.Rooms),
		Total: info.Total,
	}
}

// EndRoomRequestToModel конвертирует запрос завершения комнаты из репозитория в доменную модель.
func EndRoomRequestToModel(info repoModel.EndRoomRequest) model.EndRoomRequest {
	return model.EndRoomRequest{
		RoomUUID:  info.RoomUUID,
		OwnerUUID: info.OwnerUUID,
	}
}

// EndRoomRequestToRepoModel конвертирует доменную модель запроса завершения комнаты в модель репозитория.
func EndRoomRequestToRepoModel(info model.EndRoomRequest) repoModel.EndRoomRequest {
	return repoModel.EndRoomRequest{
		RoomUUID:  info.RoomUUID,
		OwnerUUID: info.OwnerUUID,
	}
}

// EndRoomResponseToModel конвертирует ответ завершения комнаты из репозитория в доменную модель.
func EndRoomResponseToModel(info repoModel.EndRoomResponse) model.EndRoomResponse {
	return model.EndRoomResponse{
		Success: info.Success,
	}
}

// EndRoomResponseToRepoModel конвертирует доменную модель ответа завершения комнаты в модель репозитория.
func EndRoomResponseToRepoModel(info model.EndRoomResponse) repoModel.EndRoomResponse {
	return repoModel.EndRoomResponse{
		Success: info.Success,
	}
}

