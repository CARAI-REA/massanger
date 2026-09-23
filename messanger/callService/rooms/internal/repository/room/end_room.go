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

func (r *roomRepository) EndRoom(ctx context.Context, req model.EndRoomRequest) (model.EndRoomResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.EndRoomRequestToRepoModel(req)

	builder := sq.
		Update("rooms").
		PlaceholderFormat(sq.Dollar).
		Set("status", "ended").
		Set("updated_at", time.Now()).
		Where(sq.Eq{"room_uuid": repoReq.RoomUUID, "owner_uuid": repoReq.OwnerUUID, "deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос завершения комнаты", zap.Error(err))
		return model.EndRoomResponse{}, err
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "Не удалось завершить комнату", zap.Error(err))
		return model.EndRoomResponse{}, err
	}

	if res.RowsAffected() == 0 {
		return model.EndRoomResponse{}, model.ErrRoomNotFound
	}

	return model.EndRoomResponse{Success: true}, nil
}

