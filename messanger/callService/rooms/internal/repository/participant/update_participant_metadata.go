package participant

import (
	"context"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"
	repoModel "rooms/internal/repository/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (r *participantRepository) UpdateParticipantMetadata(ctx context.Context, req model.UpdateParticipantMetadataRequest) (model.UpdateParticipantMetadataResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	repoReq := repoConverter.UpdateParticipantMetadataRequestToRepoModel(req)

	selectBuilder := sq.
		Select("participant_uuid").
		From("participants").
		Where(sq.Eq{"room_uuid": repoReq.RoomUUID, "user_uuid": repoReq.UserUUID, "left_at": nil}).
		OrderBy("joined_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос поиска participant_uuid", zap.Error(err))
		return model.UpdateParticipantMetadataResponse{}, err
	}

	var participantUUID string
	if err = r.pool.QueryRow(ctx, selectQuery, selectArgs...).Scan(&participantUUID); err != nil {
		if err == pgx.ErrNoRows {
			return model.UpdateParticipantMetadataResponse{}, model.ErrParticipantNotFound
		}
		logger.Error(ctx, "Не удалось найти участника для обновления metadata", zap.Error(err))
		return model.UpdateParticipantMetadataResponse{}, err
	}

	updateBuilder := sq.
		Update("participant_metadata").
		PlaceholderFormat(sq.Dollar).
		Set("display_name", repoReq.Metadata.DisplayName).
		Set("user_agent", repoReq.Metadata.UserAgent).
		Set("client_version", repoReq.Metadata.ClientVersion).
		Set("audio_muted", repoReq.Metadata.AudioMuted).
		Set("video_muted", repoReq.Metadata.VideoMuted).
		Set("role", repoReq.Metadata.Role).
		Where(sq.Eq{"participant_uuid": participantUUID})

	updateQuery, updateArgs, err := updateBuilder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос обновления participant_metadata", zap.Error(err))
		return model.UpdateParticipantMetadataResponse{}, err
	}

	res, err := r.pool.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		logger.Error(ctx, "Не удалось обновить participant_metadata", zap.Error(err))
		return model.UpdateParticipantMetadataResponse{}, err
	}

	if res.RowsAffected() == 0 {
		return model.UpdateParticipantMetadataResponse{}, model.ErrParticipantNotFound
	}

	repoParticipant := repoModel.Participant{
		UserUUID: repoReq.UserUUID,
		RommUUID: repoReq.RoomUUID,
		Metadata: repoReq.Metadata,
	}

	repoRes := &repoModel.UpdateParticipantMetadataResponse{
		Seccess:    true,
		Participant: repoParticipant,
	}

	return repoConverter.UpdateParticipantMetadataResponseToModel(repoRes), nil
}

