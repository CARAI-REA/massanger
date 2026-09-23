package room

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"
)

func (r *roomRepository) UpdateRoom(ctx context.Context, req model.UpdateRoomRequest) (model.UpdateRoomResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.UpdateRoomRequestToRepoModel(req)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		logger.Error(ctx, "failed to begin update room transaction", zap.Error(err))
		return model.UpdateRoomResponse{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	updateRoomBuilder := sq.
		Update("rooms").
		PlaceholderFormat(sq.Dollar).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"room_uuid": repoReq.RoomUUID, "owner_uuid": repoReq.OwnerUUID, "deleted_at": nil})

	roomQuery, roomArgs, err := updateRoomBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос к таблице rooms", zap.Error(err))
		return model.UpdateRoomResponse{}, err
	}

	roomResult, err := tx.Exec(ctx, roomQuery, roomArgs...)
	if err != nil {
		logger.Error(ctx, "Не удалось обновить запись в таблице rooms", zap.Error(err))
		return model.UpdateRoomResponse{}, err
	}
	if roomResult.RowsAffected() == 0 {
		return model.UpdateRoomResponse{}, model.ErrRoomNotFound
	}

	updateSettingsBuilder := sq.
		Update("room_settings").
		PlaceholderFormat(sq.Dollar).
		Set("max_participants", repoReq.RoomSettings.MaxParticipants).
		Set("allowed_users", repoReq.RoomSettings.AllowedUsers).
		Set("quality", repoReq.RoomSettings.Quality).
		Set("auto_close", repoReq.RoomSettings.AutoClose).
		Set("recording_enabled", repoReq.RoomSettings.RecordingEnabled).
		Where(sq.Eq{"room_uuid": repoReq.RoomUUID})

	settingsQuery, settingsArgs, err := updateSettingsBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос к таблице room_settings", zap.Error(err))
		return model.UpdateRoomResponse{}, err
	}

	settingsResult, err := tx.Exec(ctx, settingsQuery, settingsArgs...)
	if err != nil {
		logger.Error(ctx, "Не удалось обновить запись в таблице room_settings", zap.Error(err))
		return model.UpdateRoomResponse{}, err
	}
	if settingsResult.RowsAffected() == 0 {
		return model.UpdateRoomResponse{}, model.ErrRoomNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error(ctx, "failed to commit update room transaction", zap.Error(err))
		return model.UpdateRoomResponse{}, err
	}

	getRes, err := r.GetRoom(ctx, model.GetRoomRequest{RoomUUID: repoReq.RoomUUID})
	if err != nil {
		return model.UpdateRoomResponse{}, err
	}

	return model.UpdateRoomResponse{
		Success: true,
		Room:    getRes.Room,
	}, nil
}
