package converter

import (
	"rooms/internal/model"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

func ParticipantMetadataToModel(info *roomsV1.ParticipantMetadata) model.ParticipantMetadata {
	if info == nil {
		return model.ParticipantMetadata{}
	}
	return model.ParticipantMetadata{
		DisplayName:   info.DisplayName,
		UserAgent:     info.UserAgent,
		ClientVersion: info.ClientVersion,
		AudioMuted:    info.AudioMuted,
		VideoMuted:    info.VideoMuted,
		Role:          info.Role,
	}
}

func ParticipantMetadataToProto(info model.ParticipantMetadata) *roomsV1.ParticipantMetadata {
	return &roomsV1.ParticipantMetadata{
		DisplayName:   info.DisplayName,
		UserAgent:     info.UserAgent,
		ClientVersion: info.ClientVersion,
		AudioMuted:    info.AudioMuted,
		VideoMuted:    info.VideoMuted,
		Role:          info.Role,
	}
}

func ParticipantToModel(info *roomsV1.Participant) model.Participant {
	if info == nil {
		return model.Participant{}
	}
	return model.Participant{
		UserUUID: info.UserUuid,
		RommUUID: info.RoomUuid,
		JoinedAt: timestampToTimePtr(info.JoinedAt),
		LeftAt:   timestampToTimePtr(info.LeftAt),
		Metadata: ParticipantMetadataToModel(info.Metadata),
	}
}

func ParticipantToProto(info model.Participant) *roomsV1.Participant {
	return &roomsV1.Participant{
		UserUuid: info.UserUUID,
		RoomUuid: info.RommUUID,
		JoinedAt: timePtrToTimestamp(info.JoinedAt),
		LeftAt:   timePtrToTimestamp(info.LeftAt),
		Metadata: ParticipantMetadataToProto(info.Metadata),
	}
}

func AddParticipantRequestToModel(info *roomsV1.AddParticipantRequest) model.AddParticipantRequest {
	if info == nil {
		return model.AddParticipantRequest{}
	}
	return model.AddParticipantRequest{
		RoomUUID: info.RoomUuid,
		UserUUID: info.UserUuid,
		Metadata: ParticipantMetadataToModel(info.Metadata),
	}
}

func AddParticipantRequestToProto(info model.AddParticipantRequest) *roomsV1.AddParticipantRequest {
	return &roomsV1.AddParticipantRequest{
		RoomUuid: info.RoomUUID,
		UserUuid: info.UserUUID,
		Metadata: ParticipantMetadataToProto(info.Metadata),
	}
}

func AddParticipantResponseToModel(info *roomsV1.AddParticipantResponse) model.AddParticipantResponse {
	if info == nil {
		return model.AddParticipantResponse{}
	}
	return model.AddParticipantResponse{
		Success:     info.Success,
		JoinToken:   info.JoinToken,
		Participant: ParticipantToModel(info.Participant),
	}
}

func AddParticipantResponseToProto(info model.AddParticipantResponse) *roomsV1.AddParticipantResponse {
	return &roomsV1.AddParticipantResponse{
		Success:     info.Success,
		JoinToken:   info.JoinToken,
		Participant: ParticipantToProto(info.Participant),
	}
}

func RemoveParticipantRequestToModel(info *roomsV1.RemoveParticipantRequest) model.RemoveParticipantRequest {
	if info == nil {
		return model.RemoveParticipantRequest{}
	}
	return model.RemoveParticipantRequest{
		RoomUUID:  info.RoomUuid,
		UserUUID:  info.UserUuid,
		RemovedBy: info.RemovedBy,
	}
}

func RemoveParticipantRequestToProto(info model.RemoveParticipantRequest) *roomsV1.RemoveParticipantRequest {
	return &roomsV1.RemoveParticipantRequest{
		RoomUuid:  info.RoomUUID,
		UserUuid:  info.UserUUID,
		RemovedBy: info.RemovedBy,
	}
}

func RemoveParticipantResponseToModel(info *roomsV1.RemoveParticipantResponse) model.RemoveParticipantResponse {
	if info == nil {
		return model.RemoveParticipantResponse{}
	}
	return model.RemoveParticipantResponse{
		Success: info.Success,
	}
}

func RemoveParticipantResponseToProto(info model.RemoveParticipantResponse) *roomsV1.RemoveParticipantResponse {
	return &roomsV1.RemoveParticipantResponse{
		Success: info.Success,
	}
}

func ParticipantsSliceToModel(parts []*roomsV1.Participant) []model.Participant {
	result := make([]model.Participant, len(parts))
	for i, participant := range parts {
		result[i] = ParticipantToModel(participant)
	}
	return result
}

func ParticipantsSliceToProto(parts []model.Participant) []*roomsV1.Participant {
	result := make([]*roomsV1.Participant, len(parts))
	for i, participant := range parts {
		result[i] = ParticipantToProto(participant)
	}
	return result
}

func GetParticipantRequestToModel(info *roomsV1.GetParticipantsRequest) model.GetParticipantsRequest {
	if info == nil {
		return model.GetParticipantsRequest{}
	}
	return model.GetParticipantsRequest{
		RoomUUID:   info.RoomUuid,
		OnlyActive: info.OnlyActive,
	}
}

func GetParticipantsRequestToProto(info model.GetParticipantsRequest) *roomsV1.GetParticipantsRequest {
	return &roomsV1.GetParticipantsRequest{
		RoomUuid:   info.RoomUUID,
		OnlyActive: info.OnlyActive,
	}
}

func GetParticipantsResponseToModel(info *roomsV1.GetParticipantsResponse) model.GetParticipantsResponse {
	if info == nil {
		return model.GetParticipantsResponse{}
	}
	return model.GetParticipantsResponse{
		Partisipants: ParticipantsSliceToModel(info.Participants),
		ActiveCount:  info.ActiveCount,
	}
}

func GetParticipantsResponseToProto(info model.GetParticipantsResponse) *roomsV1.GetParticipantsResponse {
	return &roomsV1.GetParticipantsResponse{
		Participants: ParticipantsSliceToProto(info.Partisipants),
		ActiveCount:  info.ActiveCount,
	}
}

func IsParticipantRequestToModel(info *roomsV1.IsParticipantRequest) model.IsParticipantRequest {
	if info == nil {
		return model.IsParticipantRequest{}
	}
	return model.IsParticipantRequest{
		RoomUUID: info.RoomUuid,
		UserUUID: info.UserUuid,
	}
}

func IsParticipantRequestToProto(info model.IsParticipantRequest) *roomsV1.IsParticipantRequest {
	return &roomsV1.IsParticipantRequest{
		RoomUuid: info.RoomUUID,
		UserUuid: info.UserUUID,
	}
}

func IsParticipantResponseToModel(info *roomsV1.IsParticipantResponse) model.IsParticipantResponse {
	if info == nil {
		return model.IsParticipantResponse{}
	}
	return model.IsParticipantResponse{
		IsActive:    info.IsActive,
		Participant: ParticipantToModel(info.Participant),
	}
}

func IsParticipantResponseToProto(info model.IsParticipantResponse) *roomsV1.IsParticipantResponse {
	return &roomsV1.IsParticipantResponse{
		IsActive:    info.IsActive,
		Participant: ParticipantToProto(info.Participant),
	}
}

func UpdateParticipantMetadataRequestToModel(info *roomsV1.UpdateParticipantMetadataRequest) model.UpdateParticipantMetadataRequest {
	if info == nil {
		return model.UpdateParticipantMetadataRequest{}
	}
	return model.UpdateParticipantMetadataRequest{
		RoomUUID: info.RoomUuid,
		UserUUID: info.UserUuid,
		Metadata: ParticipantMetadataToModel(info.Metadata),
	}
}

func UpdateParticipantMetadataRequestToProto(info model.UpdateParticipantMetadataRequest) *roomsV1.UpdateParticipantMetadataRequest {
	return &roomsV1.UpdateParticipantMetadataRequest{
		RoomUuid: info.RoomUUID,
		UserUuid: info.UserUUID,
		Metadata: ParticipantMetadataToProto(info.Metadata),
	}
}

func UpdateParticipantMetadataResponseToModel(info *roomsV1.UpdateParticipantMetadataResponse) model.UpdateParticipantMetadataResponse {
	if info == nil {
		return model.UpdateParticipantMetadataResponse{}
	}
	return model.UpdateParticipantMetadataResponse{
		Seccess:     info.Success,
		Participant: ParticipantToModel(info.Participant),
	}
}

func UpdateParticipantMetadataResponseToProto(info model.UpdateParticipantMetadataResponse) *roomsV1.UpdateParticipantMetadataResponse {
	return &roomsV1.UpdateParticipantMetadataResponse{
		Success:     info.Seccess,
		Participant: ParticipantToProto(info.Participant),
	}
}
