package participant

import (
	"context"
	"time"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"
)

func (r *participantRepository) RemoveParticipant(ctx context.Context, req model.RemoveParticipantRequest) (model.RemoveParticipantResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.RemoveParticipantRequestToRepoModel(req)

	builder := sq.
		Update("participants").
		PlaceholderFormat(sq.Dollar).
		Set("left_at", time.Now()).
		Where(sq.Eq{"room_uuid": repoReq.RoomUUID, "user_uuid": repoReq.UserUUID, "left_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос удаления участника", zap.Error(err))
		return model.RemoveParticipantResponse{}, err
	}

	res, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "Не удалось выполнить мягкое удаление участника", zap.Error(err))
		return model.RemoveParticipantResponse{}, err
	}

	if res.RowsAffected() == 0 {
		return model.RemoveParticipantResponse{}, model.ErrParticipantNotFound
	}

	return model.RemoveParticipantResponse{Success: true}, nil
}

