package kafka

import (
	"context"
	"encoding/json"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"go.uber.org/zap"

	"signaling/internal/metrics"
	"signaling/internal/service"
)

const (
	eventRoomEnded         = "room_ended"
	eventRoomDeleted       = "room_deleted"
	eventParticipantLeft   = "participant_left"
)

type roomEnvelope struct {
	EventType string `json:"event_type"`
	RoomID    string `json:"room_id"`
}

type participantEnvelope struct {
	EventType string `json:"event_type"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
}

type Handler struct {
	sessions service.SessionService
}

func NewHandler(sessions service.SessionService) *Handler {
	return &Handler{sessions: sessions}
}

func (h *Handler) Handle(ctx context.Context, value []byte) error {
	var probe struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(value, &probe); err != nil {
		logger.Error(ctx, "kafka malformed json", zap.Error(err))
		return nil
	}

	metrics.KafkaEventsTotal.WithLabelValues(probe.EventType).Inc()

	switch probe.EventType {
	case eventRoomEnded, eventRoomDeleted:
		var env roomEnvelope
		if err := json.Unmarshal(value, &env); err != nil {
			logger.Error(ctx, "kafka room envelope parse failed", zap.Error(err))
			return nil
		}
		return h.sessions.CloseRoom(ctx, env.RoomID)
	case eventParticipantLeft:
		var env participantEnvelope
		if err := json.Unmarshal(value, &env); err != nil {
			logger.Error(ctx, "kafka participant envelope parse failed", zap.Error(err))
			return nil
		}
		return h.sessions.CloseUser(ctx, env.RoomID, env.UserID)
	default:
		return nil
	}
}
