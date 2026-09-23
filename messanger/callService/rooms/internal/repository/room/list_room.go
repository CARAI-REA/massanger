package room

import (
	"context"
	"time"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"
	repoModel "rooms/internal/repository/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"
)

func (r *roomRepository) ListRooms(ctx context.Context, req model.ListRoomRequest) (model.ListRoomResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repoReq := repoConverter.ListRoomRequestToRepoModel(req)

	builder := sq.
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
		Where(sq.Eq{"r.owner_uuid": repoReq.OwnerUUID})

	if repoReq.Status != "" && repoReq.Status != "all" {
		builder = builder.Where(sq.Eq{"r.status": repoReq.Status})
	}

	if repoReq.Limit > 0 {
		builder = builder.Limit(uint64(repoReq.Limit))
	}
	if repoReq.Offset > 0 {
		builder = builder.Offset(uint64(repoReq.Offset))
	}

	builder = builder.OrderBy("r.created_at DESC").PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос списка комнат", zap.Error(err))
		return model.ListRoomResponse{}, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "Не удалось получить список комнат", zap.Error(err))
		return model.ListRoomResponse{}, err
	}
	defer rows.Close()

	rooms := make([]repoModel.Room, 0)
	for rows.Next() {
		var (
			room      repoModel.Room
			createdAt time.Time
			updatedAt time.Time
			deletedAt *time.Time
		)

		err = rows.Scan(
			&room.RoomUUID,
			&room.Name,
			&room.OwnerUUID,
			&createdAt,
			&updatedAt,
			&deletedAt,
			&room.Status,
			&room.RoomSettings.MaxParticipants,
			&room.RoomSettings.Quality,
			&room.RoomSettings.AutoClose,
			&room.RoomSettings.AllowedUsers,
			&room.RoomSettings.RecordingEnabled,
		)
		if err != nil {
			logger.Error(ctx, "Не удалось прочитать строку списка комнат", zap.Error(err))
			return model.ListRoomResponse{}, err
		}

		room.CreatedAt = &createdAt
		room.UpdatedAt = &updatedAt
		room.DeletedAt = deletedAt

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		logger.Error(ctx, "Ошибка при чтении списка комнат", zap.Error(err))
		return model.ListRoomResponse{}, err
	}

	repoRes := repoModel.ListRoomResponse{
		Rooms: rooms,
		Total: int32(len(rooms)),
	}

	return repoConverter.ListRoomsResponseToModel(repoRes), nil
}

