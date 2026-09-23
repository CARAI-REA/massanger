package room

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"
)

// joinInfo адаптирует данные комнаты/владельца под tokens.JoinInfo.
type joinInfo struct {
	userUUID string
	roomUUID string
}

func (j joinInfo) GetUserUUID() string { return j.userUUID }
func (j joinInfo) GetRoomUUID() string { return j.roomUUID }

func (r *roomRepository) CreateRoom(ctx context.Context, req model.CreateRoomRequest) (model.CreateRoomResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.CreateRoomRequestToRepoModel(req)
	now := time.Now()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		logger.Error(ctx, "failed to begin create room transaction", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		roomUUID  string
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
		status    string
	)

	createBuilder := sq.
		Insert("rooms").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "owner_uuid", "created_at", "status").
		Values(repoReq.Name, repoReq.OwnerUUID, now, "active").
		Suffix("RETURNING room_uuid, created_at, updated_at, deleted_at, status")

	roomQuery, roomArgs, err := createBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось сделать запись создания в базу данных в таблицу Rooms", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	if err := tx.QueryRow(ctx, roomQuery, roomArgs...).Scan(&roomUUID, &createdAt, &updatedAt, &deletedAt, &status); err != nil {
		logger.Error(ctx, "failed to execute create room query", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	createRoomSettingsBuilder := sq.
		Insert("room_settings").
		PlaceholderFormat(sq.Dollar).
		Columns("room_uuid", "max_participants", "allowed_users", "quality", "auto_close", "recording_enabled").
		Values(
			roomUUID,
			repoReq.RoomSettings.MaxParticipants,
			repoReq.RoomSettings.AllowedUsers,
			repoReq.RoomSettings.Quality,
			repoReq.RoomSettings.AutoClose,
			repoReq.RoomSettings.RecordingEnabled,
		)

	roomSettingsQuery, roomSettingsArgs, err := createRoomSettingsBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось сделать запись в базу данных в таблицу RoomSettings", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	if _, err := tx.Exec(ctx, roomSettingsQuery, roomSettingsArgs...); err != nil {
		logger.Error(ctx, "Не удалось сделать запись в базу данных в таблицу RoomSettings", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	// Owner must be an active participant so Signaling AssertCanJoin accepts the CreateRoom join_token.
	var participantUUID string
	ownerPartBuilder := sq.
		Insert("participants").
		PlaceholderFormat(sq.Dollar).
		Columns("room_uuid", "user_uuid", "joined_at").
		Values(roomUUID, repoReq.OwnerUUID, now).
		Suffix("RETURNING participant_uuid")
	ownerPartQuery, ownerPartArgs, err := ownerPartBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "failed to build owner participant insert", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}
	if err := tx.QueryRow(ctx, ownerPartQuery, ownerPartArgs...).Scan(&participantUUID); err != nil {
		logger.Error(ctx, "failed to insert owner as participant", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	ownerMetaBuilder := sq.
		Insert("participant_metadata").
		PlaceholderFormat(sq.Dollar).
		Columns("participant_uuid", "display_name", "user_agent", "client_version", "audio_muted", "video_muted", "role").
		Values(participantUUID, "", "", "", false, false, "owner")
	ownerMetaQuery, ownerMetaArgs, err := ownerMetaBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "failed to build owner participant_metadata insert", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}
	if _, err := tx.Exec(ctx, ownerMetaQuery, ownerMetaArgs...); err != nil {
		logger.Error(ctx, "failed to insert owner participant_metadata", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error(ctx, "failed to commit create room transaction", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	if r.token == nil {
		return model.CreateRoomResponse{}, fmt.Errorf("join token service is not configured")
	}

	tokenStr, err := r.token.GenerateJoinToken(ctx, joinInfo{
		userUUID: repoReq.OwnerUUID,
		roomUUID: roomUUID,
	})
	if err != nil {
		logger.Error(ctx, "failed to generate join token", zap.Error(err))
		return model.CreateRoomResponse{}, err
	}

	return model.CreateRoomResponse{
		Room: model.Room{
			RoomUUID:     roomUUID,
			Name:         repoReq.Name,
			OwnerUUID:    repoReq.OwnerUUID,
			CratedAt:     &createdAt,
			UpdatedAt:    &updatedAt,
			DeletedAt:    deletedAt,
			RoomSettings: req.RoomSettings,
			Status:       status,
		},
		JoinToken: tokenStr,
	}, nil
}
