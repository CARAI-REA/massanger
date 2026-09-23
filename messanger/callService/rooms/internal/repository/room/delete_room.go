package room

import (
	"context"
	"time"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"
)

func (r *roomRepository) DeleteRoom(ctx context.Context, req model.DeleteRoomRequest) (model.DeleteRoomResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.DeleteRoomRequestToRepoModel(req)

	if repoReq.Permanent {
		deleteBuilder := sq.
			Delete("rooms").
			PlaceholderFormat(sq.Dollar).
			Where(sq.Eq{"room_uuid": repoReq.RoomUUID, "owner_uuid": repoReq.OwnerUUID})

		query, args, err := deleteBuilder.ToSql()
		if err != nil {
			logger.Error(ctx, "Не удалось собрать sql запрос удаления комнаты", zap.Error(err))
			return model.DeleteRoomResponse{}, err
		}

		res, err := r.pool.Exec(ctx, query, args...)
		if err != nil {
			logger.Error(ctx, "Не удалось удалить комнату", zap.Error(err))
			return model.DeleteRoomResponse{}, err
		}

		if res.RowsAffected() == 0 {
			return model.DeleteRoomResponse{}, model.ErrRoomNotFound
		}

		return model.DeleteRoomResponse{Success: true}, nil
	}

	softDeleteBuilder := sq.
		Update("rooms").
		PlaceholderFormat(sq.Dollar).
		Set("deleted_at", time.Now()).
		Set("updated_at", time.Now()).
		Set("status", "ended").
		Where(sq.Eq{"room_uuid": repoReq.RoomUUID, "owner_uuid": repoReq.OwnerUUID, "deleted_at": nil})

	query, args, err := softDeleteBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос мягкого удаления комнаты", zap.Error(err))
		return model.DeleteRoomResponse{}, err
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "Не удалось мягко удалить комнату", zap.Error(err))
		return model.DeleteRoomResponse{}, err
	}

	if res.RowsAffected() == 0 {
		return model.DeleteRoomResponse{}, model.ErrRoomNotFound
	}

	return model.DeleteRoomResponse{Success: true}, nil
}

