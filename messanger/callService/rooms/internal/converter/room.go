package converter

import (
	"time"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"rooms/internal/model"
)

func timePtrToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func timestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	return lo.ToPtr(ts.AsTime())
}

func RoomToProto(info model.Room) *roomsV1.Room {
	return &roomsV1.Room{
		RoomUuid:     info.RoomUUID,
		Name:         info.Name,
		OwnerUuid:    info.OwnerUUID,
		CreatedAt:    timePtrToTimestamp(info.CratedAt),
		UpdatedAt:    timePtrToTimestamp(info.UpdatedAt),
		DeletedAt:    timePtrToTimestamp(info.DeletedAt),
		RoomSettings: RoomSettingsToProto(info.RoomSettings),
		Status:       info.Status,
	}
}

func RoomSettingsToProto(info model.RoomSettings) *roomsV1.RoomSettings {
	return &roomsV1.RoomSettings{
		MaxParticipants:  info.MaxParticipants,
		AllowedUsers:     info.AllowedUsers,
		Quality:          info.Quality,
		AutoClose:        info.AutoClose,
		RecordingEnabled: info.RecordingEnabled,
	}
}

func RoomSettingsToModel(info *roomsV1.RoomSettings) model.RoomSettings {
	if info == nil {
		return model.RoomSettings{}
	}
	return model.RoomSettings{
		MaxParticipants:  info.MaxParticipants,
		AllowedUsers:     info.AllowedUsers,
		Quality:          info.Quality,
		AutoClose:        info.AutoClose,
		RecordingEnabled: info.RecordingEnabled,
	}
}

func RoomToModel(info *roomsV1.Room) model.Room {
	if info == nil {
		return model.Room{}
	}
	return model.Room{
		RoomUUID:     info.RoomUuid,
		Name:         info.Name,
		OwnerUUID:    info.OwnerUuid,
		CratedAt:     timestampToTimePtr(info.CreatedAt),
		UpdatedAt:    timestampToTimePtr(info.UpdatedAt),
		DeletedAt:    timestampToTimePtr(info.DeletedAt),
		RoomSettings: RoomSettingsToModel(info.RoomSettings),
		Status:       info.Status,
	}
}

func CreateRoomRequestToModel(info *roomsV1.CreateRoomRequest) model.CreateRoomRequest {
	if info == nil {
		return model.CreateRoomRequest{}
	}
	return model.CreateRoomRequest{
		Name:         info.Name,
		OwnerUUID:    info.OwnerUuid,
		RoomSettings: RoomSettingsToModel(info.RoomSettings),
	}
}

func CreateRoomRequestToProto(info model.CreateRoomRequest) *roomsV1.CreateRoomRequest {
	return &roomsV1.CreateRoomRequest{
		Name:         info.Name,
		OwnerUuid:    info.OwnerUUID,
		RoomSettings: RoomSettingsToProto(info.RoomSettings),
	}
}

func CreateRoomResponseToModel(info *roomsV1.CreateRoomResponse) model.CreateRoomResponse {
	if info == nil {
		return model.CreateRoomResponse{}
	}
	return model.CreateRoomResponse{
		Room:      RoomToModel(info.Room),
		JoinToken: info.JoinToken,
	}
}

func CreateRoomResponseToProto(info model.CreateRoomResponse) *roomsV1.CreateRoomResponse {
	return &roomsV1.CreateRoomResponse{
		Room:      RoomToProto(info.Room),
		JoinToken: info.JoinToken,
	}
}

func GetRoomRequestToModel(info *roomsV1.GetRoomRequest) model.GetRoomRequest {
	if info == nil {
		return model.GetRoomRequest{}
	}
	return model.GetRoomRequest{
		RoomUUID: info.RoomUuid,
	}
}

func GetRoomRequestToProto(info model.GetRoomRequest) *roomsV1.GetRoomRequest {
	return &roomsV1.GetRoomRequest{
		RoomUuid: info.RoomUUID,
	}
}

func GetRoomResponseToModel(info *roomsV1.GetRoomResponse) model.GetRoomResponse {
	if info == nil {
		return model.GetRoomResponse{}
	}
	return model.GetRoomResponse{
		Room: RoomToModel(info.Room),
	}
}

func GetRoomResponseToProto(info model.GetRoomResponse) *roomsV1.GetRoomResponse {
	return &roomsV1.GetRoomResponse{
		Room: RoomToProto(info.Room),
	}
}

func UpdateRoomRequestToModel(info *roomsV1.UpdateRoomRequest) model.UpdateRoomRequest {
	if info == nil {
		return model.UpdateRoomRequest{}
	}
	return model.UpdateRoomRequest{
		RoomUUID:     info.RoomUuid,
		OwnerUUID:    info.OwnerUuid,
		RoomSettings: RoomSettingsToModel(info.RoomSettings),
	}
}

func UpdateRoomRequestToProto(info model.UpdateRoomRequest) *roomsV1.UpdateRoomRequest {
	return &roomsV1.UpdateRoomRequest{
		RoomUuid:     info.RoomUUID,
		OwnerUuid:    info.OwnerUUID,
		RoomSettings: RoomSettingsToProto(info.RoomSettings),
	}
}

func UpdateRoomResponseToModel(info *roomsV1.UpdateRoomResponse) model.UpdateRoomResponse {
	if info == nil {
		return model.UpdateRoomResponse{}
	}
	return model.UpdateRoomResponse{
		Success: info.Success,
		Room:    RoomToModel(info.Room),
	}
}

func UpdateRoomResponseToProto(info model.UpdateRoomResponse) *roomsV1.UpdateRoomResponse {
	return &roomsV1.UpdateRoomResponse{
		Success: info.Success,
		Room:    RoomToProto(info.Room),
	}
}

func DeleteRoomRequestToModel(info *roomsV1.DeleteRoomRequest) model.DeleteRoomRequest {
	if info == nil {
		return model.DeleteRoomRequest{}
	}
	return model.DeleteRoomRequest{
		RoomUUID:  info.RoomUuid,
		OwnerUUID: info.OwnerUuid,
		Permanent: info.Permanent,
	}
}

func DeleteRoomRequestToProto(info model.DeleteRoomRequest) *roomsV1.DeleteRoomRequest {
	return &roomsV1.DeleteRoomRequest{
		RoomUuid:  info.RoomUUID,
		OwnerUuid: info.OwnerUUID,
		Permanent: info.Permanent,
	}
}

func DeleteRoomResponseToModel(info *roomsV1.DeleteRoomResponse) model.DeleteRoomResponse {
	if info == nil {
		return model.DeleteRoomResponse{}
	}
	return model.DeleteRoomResponse{
		Success: info.Success,
	}
}

func DeleteRoomResponseToProto(info model.DeleteRoomResponse) *roomsV1.DeleteRoomResponse {
	return &roomsV1.DeleteRoomResponse{
		Success: info.Success,
	}
}

func ListRoomRequestToModel(info *roomsV1.ListRoomsRequest) model.ListRoomRequest {
	if info == nil {
		return model.ListRoomRequest{}
	}
	return model.ListRoomRequest{
		OwnerUUID: info.OwnerUuid,
		Status:    info.Status,
		Limit:     info.Limit,
		Offset:    info.Offset,
	}
}

func ListRoomRequestToProto(info model.ListRoomRequest) *roomsV1.ListRoomsRequest {
	return &roomsV1.ListRoomsRequest{
		OwnerUuid: info.OwnerUUID,
		Status:    info.Status,
		Limit:     info.Limit,
		Offset:    info.Offset,
	}
}

func RoomsSliceToModel(info []*roomsV1.Room) []model.Room {
	result := make([]model.Room, len(info))
	for i, room := range info {
		result[i] = RoomToModel(room)
	}
	return result
}

func RoomsSliceToProto(info []model.Room) []*roomsV1.Room {
	result := make([]*roomsV1.Room, len(info))
	for i, room := range info {
		result[i] = RoomToProto(room)
	}
	return result
}

func ListRoomsResponseToModel(info *roomsV1.ListRoomsResponse) model.ListRoomResponse {
	if info == nil {
		return model.ListRoomResponse{}
	}
	return model.ListRoomResponse{
		Rooms: RoomsSliceToModel(info.Rooms),
		Total: info.Total,
	}
}

func ListRoomsResponseToProto(info model.ListRoomResponse) *roomsV1.ListRoomsResponse {
	return &roomsV1.ListRoomsResponse{
		Rooms: RoomsSliceToProto(info.Rooms),
		Total: info.Total,
	}
}

func EndRoomRequestToModel(info *roomsV1.EndRoomRequest) model.EndRoomRequest {
	if info == nil {
		return model.EndRoomRequest{}
	}
	return model.EndRoomRequest{
		RoomUUID:  info.RoomUuid,
		OwnerUUID: info.OwnerUuid,
	}
}

func EndRoomRequestToProto(info model.EndRoomRequest) *roomsV1.EndRoomRequest {
	return &roomsV1.EndRoomRequest{
		RoomUuid:  info.RoomUUID,
		OwnerUuid: info.OwnerUUID,
	}
}

func EndRoomResponseToModel(info *roomsV1.EndRoomResponse) model.EndRoomResponse {
	if info == nil {
		return model.EndRoomResponse{}
	}
	return model.EndRoomResponse{
		Success: info.Success,
	}
}

func EndRoomResponseToProto(info model.EndRoomResponse) *roomsV1.EndRoomResponse {
	return &roomsV1.EndRoomResponse{
		Success: info.Success,
	}
}

func AssertCanJoinRequestToModel(info *roomsV1.AssertCanJoinRequest) model.AssertCanJoinRequest {
	if info == nil {
		return model.AssertCanJoinRequest{}
	}
	return model.AssertCanJoinRequest{
		RoomUUID: info.RoomUuid,
		UserUUID: info.UserUuid,
	}
}

func AssertCanJoinResponseToProto(info model.AssertCanJoinResponse) *roomsV1.AssertCanJoinResponse {
	return &roomsV1.AssertCanJoinResponse{
		Ok:               info.Ok,
		RoomStatus:       info.RoomStatus,
		RecordingEnabled: info.RecordingEnabled,
	}
}

func RefreshJoinTokenRequestToModel(info *roomsV1.RefreshJoinTokenRequest) model.RefreshJoinTokenRequest {
	if info == nil {
		return model.RefreshJoinTokenRequest{}
	}
	return model.RefreshJoinTokenRequest{
		RoomUUID: info.RoomUuid,
		UserUUID: info.UserUuid,
	}
}

func RefreshJoinTokenResponseToProto(info model.RefreshJoinTokenResponse) *roomsV1.RefreshJoinTokenResponse {
	return &roomsV1.RefreshJoinTokenResponse{
		JoinToken: info.JoinToken,
	}
}

func GetTURNCredentialsRequestToModel(info *roomsV1.GetTURNCredentialsRequest) model.GetTURNCredentialsRequest {
	if info == nil {
		return model.GetTURNCredentialsRequest{}
	}
	return model.GetTURNCredentialsRequest{
		RoomUUID: info.RoomUuid,
		UserUUID: info.UserUuid,
	}
}

func GetTURNCredentialsResponseToProto(info model.GetTURNCredentialsResponse) *roomsV1.GetTURNCredentialsResponse {
	servers := make([]*roomsV1.ICEServer, len(info.ICEServers))
	for i, s := range info.ICEServers {
		servers[i] = &roomsV1.ICEServer{
			Urls:       s.URLs,
			Username:   s.Username,
			Credential: s.Credential,
		}
	}
	return &roomsV1.GetTURNCredentialsResponse{
		IceServers: servers,
		TtlSeconds: info.TTLSeconds,
	}
}
