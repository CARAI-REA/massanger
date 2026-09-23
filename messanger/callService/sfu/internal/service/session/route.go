package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
)

func (s *Service) Handle(ctx context.Context, roomUUID, userUUID string, msg model.Envelope) error {
	metrics.MessagesTotal.WithLabelValues(string(msg.Type), "in").Inc()
	_ = s.affinity.Refresh(ctx, roomUUID, s.instanceID, s.affinityTTL)

	switch msg.Type {
	case model.TypePing:
		if p, ok := s.rooms.Get(roomUUID, userUUID); ok && p.Conn != nil {
			_ = p.Conn.Send(ctx, model.Envelope{
				Type:     model.TypePong,
				RoomUUID: roomUUID,
				Payload:  json.RawMessage(`{}`),
				TS:       time.Now().UTC().Format(time.RFC3339Nano),
			})
			metrics.MessagesTotal.WithLabelValues(string(model.TypePong), "out").Inc()
		}
		return nil
	case model.TypeBye:
		if p, ok := s.rooms.Get(roomUUID, userUUID); ok {
			return s.Disconnect(ctx, p.Conn, roomUUID, userUUID, "bye")
		}
		return nil
	case model.TypeOffer, model.TypeAnswer, model.TypeICECandidate:
		err := s.rooms.HandleSignal(roomUUID, userUUID, msg)
		if err != nil {
			code := model.CodeInvalidArgument
			if errors.Is(err, model.ErrNotFound) {
				code = model.CodeNotFound
			}
			return s.sendError(ctx, roomUUID, userUUID, code, err.Error())
		}
		if msg.Type == model.TypeOffer {
			metrics.MessagesTotal.WithLabelValues(string(model.TypeAnswer), "out").Inc()
		}
		return nil
	case model.TypePong:
		return nil
	default:
		return s.sendError(ctx, roomUUID, userUUID, model.CodeInvalidArgument, "unsupported message type")
	}
}

func (s *Service) sendError(ctx context.Context, roomUUID, userUUID, code, message string) error {
	metrics.MessageErrorsTotal.WithLabelValues(code).Inc()
	if p, ok := s.rooms.Get(roomUUID, userUUID); ok && p.Conn != nil {
		_ = p.Conn.Send(ctx, converter.ErrorEnvelope(roomUUID, code, message))
		metrics.MessagesTotal.WithLabelValues(string(model.TypeError), "out").Inc()
	}
	return nil
}
