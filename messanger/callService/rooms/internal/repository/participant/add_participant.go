package participant

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

type joinInfo struct {
	userUUID string
	roomUUID string
}

func (j joinInfo) GetUserUUID() string { return j.userUUID }
func (j joinInfo) GetRoomUUID() string { return j.roomUUID }

func (r *participantRepository) AddParticipant(ctx context.Context, req model.AddParticipantRequest) (model.AddParticipantResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.AddParticipantRequestToRepoModel(req)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		logger.Error(ctx, "failed to begin add participant transaction", zap.Error(err))
		return model.AddParticipantResponse{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		participantUUID string
		joinedAt        time.Time
		leftAt          *time.Time
	)

	builder := sq.
		Insert("participants").
		PlaceholderFormat(sq.Dollar).
		Columns("room_uuid", "user_uuid", "joined_at").
		Values(repoReq.RoomUUID, repoReq.UserUUID, time.Now()).
		Suffix("RETURNING participant_uuid, joined_at, left_at")

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос добавления участника", zap.Error(err))
		return model.AddParticipantResponse{}, err
	}

	if err = tx.QueryRow(ctx, query, args...).Scan(&participantUUID, &joinedAt, &leftAt); err != nil {
		logger.Error(ctx, "Не удалось добавить участника в таблицу participants", zap.Error(err))
		return model.AddParticipantResponse{}, err
	}

	metaBuilder := sq.
		Insert("participant_metadata").
		PlaceholderFormat(sq.Dollar).
		Columns("participant_uuid", "display_name", "user_agent", "client_version", "audio_muted", "video_muted", "role").
		Values(
			participantUUID,
			repoReq.Metadata.DisplayName,
			repoReq.Metadata.UserAgent,
			repoReq.Metadata.ClientVersion,
			repoReq.Metadata.AudioMuted,
			repoReq.Metadata.VideoMuted,
			repoReq.Metadata.Role,
		)

	metaQuery, metaArgs, err := metaBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос добавления participant_metadata", zap.Error(err))
		return model.AddParticipantResponse{}, err
	}

	if _, err = tx.Exec(ctx, metaQuery, metaArgs...); err != nil {
		logger.Error(ctx, "Не удалось добавить participant_metadata", zap.Error(err))
		return model.AddParticipantResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error(ctx, "failed to commit add participant transaction", zap.Error(err))
		return model.AddParticipantResponse{}, err
	}

	if r.token == nil {
		return model.AddParticipantResponse{}, fmt.Errorf("join token service is not configured")
	}

	tokenStr, tokenErr := r.token.GenerateJoinToken(ctx, joinInfo{
		userUUID: repoReq.UserUUID,
		roomUUID: repoReq.RoomUUID,
	})
	if tokenErr != nil {
		logger.Error(ctx, "Не удалось сгенерировать join token для участника", zap.Error(tokenErr))
		return model.AddParticipantResponse{}, tokenErr
	}

	return model.AddParticipantResponse{
		Success:   true,
		JoinToken: tokenStr,
		Participant: model.Participant{
			UserUUID: repoReq.UserUUID,
			RommUUID: repoReq.RoomUUID,
			JoinedAt: &joinedAt,
			LeftAt:   leftAt,
			Metadata: req.Metadata,
		},
	}, nil
}
