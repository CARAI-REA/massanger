package participant

import (
	"context"
	"time"

	"rooms/internal/model"
	repoConverter "rooms/internal/repository/converter"
	repoModel "rooms/internal/repository/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func (r *participantRepository) IsParticipant(ctx context.Context, req model.IsParticipantRequest) (model.IsParticipantResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repoReq := repoConverter.IsParticipantRequestToRepoModel(req)

	builder := sq.
		Select(
			"p.room_uuid",
			"p.user_uuid",
			"p.joined_at",
			"p.left_at",
			"pm.display_name",
			"pm.user_agent",
			"pm.client_version",
			"pm.audio_muted",
			"pm.video_muted",
			"pm.role",
		).
		From("participants p").
		LeftJoin("participant_metadata pm ON pm.participant_uuid = p.participant_uuid").
		Where(sq.Eq{"p.room_uuid": repoReq.RoomUUID, "p.user_uuid": repoReq.UserUUID, "p.left_at": nil}).
		OrderBy("p.joined_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос проверки участника", zap.Error(err))
		return model.IsParticipantResponse{}, err
	}

	var (
		part    repoModel.Participant
		joined  time.Time
		left    *time.Time
		meta    repoModel.ParticipantMetadata
	)

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&part.RommUUID,
		&part.UserUUID,
		&joined,
		&left,
		&meta.DisplayName,
		&meta.UserAgent,
		&meta.ClientVersion,
		&meta.AudioMuted,
		&meta.VideoMuted,
		&meta.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.IsParticipantResponse{IsActive: false}, nil
		}
		logger.Error(ctx, "Не удалось проверить участника", zap.Error(err))
		return model.IsParticipantResponse{}, err
	}

	part.JoinedAt = &joined
	part.LeftAt = left
	part.Metadata = meta

	repoRes := &repoModel.IsParticipantResponse{
		IsActive:    true,
		Participant: part,
	}

	return repoConverter.IsParticipantResponseToModel(repoRes), nil
}

