package participant

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

func (r *participantRepository) GetParticipant(ctx context.Context, req model.GetParticipantsRequest) (model.GetParticipantsResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repoReq := repoConverter.GetParticipantsRequestToRepoModel(req)

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
		Where(sq.Eq{"p.room_uuid": repoReq.RoomUUID})

	if repoReq.OnlyActive {
		builder = builder.Where(sq.Eq{"p.left_at": nil})
	}

	builder = builder.OrderBy("p.joined_at ASC").PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		logger.Error(ctx, "Не удалось собрать sql запрос списка участников", zap.Error(err))
		return model.GetParticipantsResponse{}, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		logger.Error(ctx, "Не удалось получить список участников", zap.Error(err))
		return model.GetParticipantsResponse{}, err
	}
	defer rows.Close()

	parts := make([]repoModel.Participant, 0)
	var activeCount int32

	for rows.Next() {
		var (
			p       repoModel.Participant
			joined  time.Time
			left    *time.Time
			meta    repoModel.ParticipantMetadata
		)

		err = rows.Scan(
			&p.RommUUID,
			&p.UserUUID,
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
			logger.Error(ctx, "Не удалось прочитать строку участника", zap.Error(err))
			return model.GetParticipantsResponse{}, err
		}

		p.JoinedAt = &joined
		p.LeftAt = left
		p.Metadata = meta
		if p.LeftAt == nil {
			activeCount++
		}
		parts = append(parts, p)
	}

	if err = rows.Err(); err != nil {
		logger.Error(ctx, "Ошибка чтения rows участников", zap.Error(err))
		return model.GetParticipantsResponse{}, err
	}

	repoRes := &repoModel.GetParticipantsResponse{
		Partisipants: parts,
		ActiveCount:  activeCount,
	}

	return repoConverter.GetParticipantsResponseToModel(repoRes), nil
}

