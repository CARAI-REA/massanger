package room

import (
	"context"
	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"
	repoModel "rooms/internal/repository/model"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (r *roomRepository) GetRoom(ctx context.Context, req model.GetRoomRequest) (model.GetRoomResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	getRoomBuilder := sq.
		Select(
			"r.room_uuid",
			"r.name",
			"r.owner_uuid",
			"r.created_at",
			"r.updated_at",
			"r.deleted_at",
			"r.status",
			"rs.max_participants",
			"rs.quality",
			"rs.auto_close",
			"rs.allowed_users",
			"rs.recording_enabled",
		).
		From("rooms r").
		Join("room_settings rs ON rs.room_uuid = r.room_uuid").
		Where(sq.Eq{"r.room_uuid": req.RoomUUID}).
		PlaceholderFormat(sq.Dollar)

	roomQuery, roomArgs, err := getRoomBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос к базе данных к таблице rooms", zap.Error(err))
		return model.GetRoomResponse{}, err
	}

	var (
		res       repoModel.GetRoomResponse
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	err = r.pool.QueryRow(ctx, roomQuery, roomArgs...).Scan(
		&res.Room.RoomUUID,
		&res.Room.Name,
		&res.Room.OwnerUUID,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&res.Room.Status,
		&res.Room.RoomSettings.MaxParticipants,
		&res.Room.RoomSettings.Quality,
		&res.Room.RoomSettings.AutoClose,
		&res.Room.RoomSettings.AllowedUsers,
		&res.Room.RoomSettings.RecordingEnabled,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.GetRoomResponse{}, model.ErrRoomNotFound
		}
		logger.Error(ctx, "Не удалось получить запись из таблиц rooms и room_settings", zap.Error(err))
		return model.GetRoomResponse{}, err
	}

	res.Room.CreatedAt = &createdAt
	res.Room.UpdatedAt = &updatedAt
	res.Room.DeletedAt = deletedAt

	return repoConverter.GetRoomResponseToModel(res), nil
}
