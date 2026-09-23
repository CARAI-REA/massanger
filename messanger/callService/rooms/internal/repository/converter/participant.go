package converter

import (
	model "rooms/internal/model"
	repoModel "rooms/internal/repository/model"
)

// ParticipantMetadataToModel конвертирует метаданные участника из модели репозитория в доменную модель.
func ParticipantMetadataToModel(info *repoModel.ParticipantMetadata) model.ParticipantMetadata {
	return model.ParticipantMetadata{
		DisplayName:   info.DisplayName,
		UserAgent:     info.UserAgent,
		ClientVersion: info.ClientVersion,
		AudioMuted:    info.AudioMuted,
		VideoMuted:    info.VideoMuted,
		Role:          info.Role,
	}
}

// ParticipantMetadataToRepoModel конвертирует доменную модель метаданных участника в модель репозитория.
func ParticipantMetadataToRepoModel(info model.ParticipantMetadata) *repoModel.ParticipantMetadata {
	return &repoModel.ParticipantMetadata{
		DisplayName:   info.DisplayName,
		UserAgent:     info.UserAgent,
		ClientVersion: info.ClientVersion,
		AudioMuted:    info.AudioMuted,
		VideoMuted:    info.VideoMuted,
		Role:          info.Role,
	}
}

// ParticipantToModel конвертирует участника из модели репозитория в доменную модель.
func ParticipantToModel(info *repoModel.Participant) model.Participant {
	return model.Participant{
		UserUUID: info.UserUUID,
		RommUUID: info.RommUUID,
		JoinedAt: info.JoinedAt,
		LeftAt:   info.LeftAt,
		Metadata: ParticipantMetadataToModel(&info.Metadata),
	}
}

// ParticipantToRepoModel конвертирует доменную модель участника в модель репозитория.
func ParticipantToRepoModel(info model.Participant) *repoModel.Participant {
	return &repoModel.Participant{
		UserUUID: info.UserUUID,
		RommUUID: info.RommUUID,
		JoinedAt: info.JoinedAt,
		LeftAt:   info.LeftAt,
		Metadata: *ParticipantMetadataToRepoModel(info.Metadata),
	}
}

// AddParticipantRequestToModel конвертирует запрос добавления участника из репозитория в доменную модель.
func AddParticipantRequestToModel(info *repoModel.AddParticipantRequest) model.AddParticipantRequest {
	return model.AddParticipantRequest{
		RoomUUID: info.RoomUUID,
		UserUUID: info.UserUUID,
		Metadata: ParticipantMetadataToModel(&info.Metadata),
	}
}

// AddParticipantRequestToRepoModel конвертирует доменную модель запроса добавления участника в модель репозитория.
func AddParticipantRequestToRepoModel(info model.AddParticipantRequest) *repoModel.AddParticipantRequest {
	return &repoModel.AddParticipantRequest{
		RoomUUID: info.RoomUUID,
		UserUUID: info.UserUUID,
		Metadata: *ParticipantMetadataToRepoModel(info.Metadata),
	}
}

// AddParticipantResponseToModel конвертирует ответ добавления участника из репозитория в доменную модель.
func AddParticipantResponseToModel(info *repoModel.AddParticipantResponse) model.AddParticipantResponse {
	return model.AddParticipantResponse{
		Success:     info.Success,
		JoinToken:   info.JoinToken,
		Participant: ParticipantToModel(&info.Participant),
	}
}

// AddParticipantResponseToRepoModel конвертирует доменную модель ответа добавления участника в модель репозитория.
func AddParticipantResponseToRepoModel(info model.AddParticipantResponse) *repoModel.AddParticipantResponse {
	return &repoModel.AddParticipantResponse{
		Success:     info.Success,
		JoinToken:   info.JoinToken,
		Participant: *ParticipantToRepoModel(info.Participant),
	}
}

// RemoveParticipantRequestToModel конвертирует запрос удаления участника из репозитория в доменную модель.
func RemoveParticipantRequestToModel(info *repoModel.RemoveParticipantRequest) model.RemoveParticipantRequest {
	return model.RemoveParticipantRequest{
		RoomUUID:  info.RoomUUID,
		UserUUID:  info.UserUUID,
		RemovedBy: info.RemovedBy,
	}
}

// RemoveParticipantRequestToRepoModel конвертирует доменную модель запроса удаления участника в модель репозитория.
func RemoveParticipantRequestToRepoModel(info model.RemoveParticipantRequest) *repoModel.RemoveParticipantRequest {
	return &repoModel.RemoveParticipantRequest{
		RoomUUID:  info.RoomUUID,
		UserUUID:  info.UserUUID,
		RemovedBy: info.RemovedBy,
	}
}

// RemoveParticipantResponseToModel конвертирует ответ удаления участника из репозитория в доменную модель.
func RemoveParticipantResponseToModel(info *repoModel.RemoveParticipantResponse) model.RemoveParticipantResponse {
	return model.RemoveParticipantResponse{
		Success: info.Success,
	}
}

// RemoveParticipantResponseToRepoModel конвертирует доменную модель ответа удаления участника в модель репозитория.
func RemoveParticipantResponseToRepoModel(info model.RemoveParticipantResponse) *repoModel.RemoveParticipantResponse {
	return &repoModel.RemoveParticipantResponse{
		Success: info.Success,
	}
}

// ParticipantsSliseToModel конвертирует слайс участников из модели репозитория в доменную модель.
func ParticipantsSliseToModel(parts []*repoModel.Participant) []model.Participant {
	result := make([]model.Participant, len(parts))

	for i, participant := range parts {
		result[i] = ParticipantToModel(participant)
	}

	return result
}

// ParticipantsSliseToRepoModel конвертирует слайс доменных участников в слайс моделей репозитория.
func ParticipantsSliseToRepoModel(parts []model.Participant) []*repoModel.Participant {
	result := make([]*repoModel.Participant, len(parts))

	for i, participant := range parts {
		result[i] = ParticipantToRepoModel(participant)
	}

	return result
}

// GetParticipantRequestToModel конвертирует запрос получения участников из репозитория в доменную модель.
func GetParticipantRequestToModel(info *repoModel.GetParticipantsRequest) model.GetParticipantsRequest {
	return model.GetParticipantsRequest{
		RoomUUID:  info.RoomUUID,
		OnlyActive: info.OnlyActive,
	}
}

// GetParticipantsRequestToRepoModel конвертирует доменную модель запроса получения участников в модель репозитория.
func GetParticipantsRequestToRepoModel(info model.GetParticipantsRequest) *repoModel.GetParticipantsRequest {
	return &repoModel.GetParticipantsRequest{
		RoomUUID:  info.RoomUUID,
		OnlyActive: info.OnlyActive,
	}
}

// GetParticipantsResponseToModel конвертирует ответ получения участников из репозитория в доменную модель.
func GetParticipantsResponseToModel(info *repoModel.GetParticipantsResponse) model.GetParticipantsResponse {
	return model.GetParticipantsResponse{
		Partisipants: ParticipantsSliseToModel(ptrSliceParticipants(info.Partisipants)),
		ActiveCount:  info.ActiveCount,
	}
}

// GetParticipantsResponseToRepoModel конвертирует доменную модель ответа получения участников в модель репозитория.
func GetParticipantsResponseToRepoModel(info model.GetParticipantsResponse) *repoModel.GetParticipantsResponse {
	return &repoModel.GetParticipantsResponse{
		Partisipants: derefSliceParticipants(ParticipantsSliseToRepoModel(info.Partisipants)),
		ActiveCount:  info.ActiveCount,
	}
}

// IsParticipantRequestToModel конвертирует запрос проверки участника из репозитория в доменную модель.
func IsParticipantRequestToModel(info *repoModel.IsParticipantRequest) model.IsParticipantRequest {
	return model.IsParticipantRequest{
		RoomUUID: info.RoomUUID,
		UserUUID: info.UserUUID,
	}
}

// IsParticipantRequestToRepoModel конвертирует доменную модель запроса проверки участника в модель репозитория.
func IsParticipantRequestToRepoModel(info model.IsParticipantRequest) *repoModel.IsParticipantRequest {
	return &repoModel.IsParticipantRequest{
		RoomUUID: info.RoomUUID,
		UserUUID: info.UserUUID,
	}
}

// IsParticipantResponseToModel конвертирует ответ проверки участника из репозитория в доменную модель.
func IsParticipantResponseToModel(info *repoModel.IsParticipantResponse) model.IsParticipantResponse {
	return model.IsParticipantResponse{
		IsActive:    info.IsActive,
		Participant: ParticipantToModel(&info.Participant),
	}
}

// IsParticipantResponseToRepoModel конвертирует доменную модель ответа проверки участника в модель репозитория.
func IsParticipantResponseToRepoModel(info model.IsParticipantResponse) *repoModel.IsParticipantResponse {
	return &repoModel.IsParticipantResponse{
		IsActive:    info.IsActive,
		Participant: *ParticipantToRepoModel(info.Participant),
	}
}

// UpdateParticipantMetadataRequestToModel конвертирует запрос обновления метаданных участника из репозитория в доменную модель.
func UpdateParticipantMetadataRequestToModel(info *repoModel.UpdateParticipantMetadataRequest) model.UpdateParticipantMetadataRequest {
	return model.UpdateParticipantMetadataRequest{
		RoomUUID: info.RoomUUID,
		UserUUID: info.UserUUID,
		Metadata: ParticipantMetadataToModel(&info.Metadata),
	}
}

// UpdateParticipantMetadataRequestToRepoModel конвертирует доменную модель запроса обновления метаданных участника в модель репозитория.
func UpdateParticipantMetadataRequestToRepoModel(info model.UpdateParticipantMetadataRequest) *repoModel.UpdateParticipantMetadataRequest {
	return &repoModel.UpdateParticipantMetadataRequest{
		RoomUUID: info.RoomUUID,
		UserUUID: info.UserUUID,
		Metadata: *ParticipantMetadataToRepoModel(info.Metadata),
	}
}

// UpdateParticipantMetadataResponseToModel конвертирует ответ обновления метаданных участника из репозитория в доменную модель.
func UpdateParticipantMetadataResponseToModel(info *repoModel.UpdateParticipantMetadataResponse) model.UpdateParticipantMetadataResponse {
	return model.UpdateParticipantMetadataResponse{
		Seccess:    info.Seccess,
		Participant: ParticipantToModel(&info.Participant),
	}
}

// UpdateParticipantMetadataResponseToRepoModel конвертирует доменную модель ответа обновления метаданных участника в модель репозитория.
func UpdateParticipantMetadataResponseToRepoModel(info model.UpdateParticipantMetadataResponse) *repoModel.UpdateParticipantMetadataResponse {
	return &repoModel.UpdateParticipantMetadataResponse{
		Seccess:    info.Seccess,
		Participant: *ParticipantToRepoModel(info.Participant),
	}
}

// вспомогательные функции для конвертации []Participant <-> []*Participant
func ptrSliceParticipants(in []repoModel.Participant) []*repoModel.Participant {
	out := make([]*repoModel.Participant, len(in))
	for i := range in {
		out[i] = &in[i]
	}
	return out
}

func derefSliceParticipants(in []*repoModel.Participant) []repoModel.Participant {
	out := make([]repoModel.Participant, len(in))
	for i, p := range in {
		out[i] = *p
	}
	return out
}
