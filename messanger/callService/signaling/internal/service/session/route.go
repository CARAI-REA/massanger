package session

import (
	"context"
	"encoding/json"
	"time"

	"signaling/internal/converter"
	"signaling/internal/metrics"
	"signaling/internal/model"
)

func (s *Service) Handle(ctx context.Context, roomUUID, userUUID string, msg model.Envelope) error {
	metrics.MessagesTotal.WithLabelValues(string(msg.Type), "in").Inc()
	_ = s.presence.RefreshTTL(ctx, roomUUID, userUUID)

	switch msg.Type {
	case model.TypePing:
		pong := model.Envelope{
			Type:     model.TypePong,
			RoomUUID: roomUUID,
			Payload:  json.RawMessage(`{}`),
			TS:       time.Now().UTC().Format(time.RFC3339Nano),
		}
		if c, ok := s.hub.Get(roomUUID, userUUID); ok {
			_ = c.Send(ctx, pong)
			metrics.MessagesTotal.WithLabelValues(string(model.TypePong), "out").Inc()
		}
		return nil
	case model.TypeBye:
		if c, ok := s.hub.Get(roomUUID, userUUID); ok {
			return s.Disconnect(ctx, c, roomUUID, userUUID, "bye")
		}
		return nil
	case model.TypeOffer, model.TypeAnswer, model.TypeICECandidate:
		return s.routeSignal(ctx, roomUUID, userUUID, msg)
	case model.TypePong:
		return nil
	default:
		return s.sendError(ctx, roomUUID, userUUID, model.CodeInvalidArgument, "unsupported message type")
	}
}

func (s *Service) routeSignal(ctx context.Context, roomUUID, userUUID string, msg model.Envelope) error {
	if msg.ToUserUUID == "" {
		return s.sendError(ctx, roomUUID, userUUID, model.CodeInvalidArgument, "to_user_uuid required")
	}
	if msg.ToUserUUID == userUUID {
		return s.sendError(ctx, roomUUID, userUUID, model.CodeInvalidArgument, "cannot send to self")
	}
	if err := validateSignalPayload(msg); err != nil {
		return s.sendError(ctx, roomUUID, userUUID, model.CodeInvalidArgument, err.Error())
	}

	msg.RoomUUID = roomUUID
	msg.FromUserUUID = userUUID
	if msg.TS == "" {
		msg.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}

	return s.deliver(ctx, roomUUID, userUUID, msg)
}

func validateSignalPayload(msg model.Envelope) error {
	switch msg.Type {
	case model.TypeOffer, model.TypeAnswer:
		var p model.SDPPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil || p.SDP == "" {
			return model.ErrInvalidArgument
		}
	case model.TypeICECandidate:
		var p model.ICEPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil || p.Candidate == "" {
			return model.ErrInvalidArgument
		}
	}
	return nil
}

func (s *Service) deliver(ctx context.Context, roomUUID, fromUser string, msg model.Envelope) error {
	if c, ok := s.hub.Get(roomUUID, msg.ToUserUUID); ok {
		if err := c.Send(ctx, msg); err != nil {
			return err
		}
		metrics.RouteLocalTotal.Inc()
		metrics.MessagesTotal.WithLabelValues(string(msg.Type), "out").Inc()
		return nil
	}

	_, ok, err := s.presence.GetInstance(ctx, roomUUID, msg.ToUserUUID)
	if err != nil {
		return s.sendError(ctx, roomUUID, fromUser, model.CodeInternal, "presence lookup failed")
	}
	if !ok {
		return s.sendError(ctx, roomUUID, fromUser, model.CodeNotFound, "peer offline")
	}

	if err := s.presence.Publish(ctx, roomUUID, msg); err != nil {
		return s.sendError(ctx, roomUUID, fromUser, model.CodeInternal, "publish failed")
	}
	metrics.RouteRemoteTotal.Inc()
	metrics.MessagesTotal.WithLabelValues(string(msg.Type), "out").Inc()
	return nil
}

func (s *Service) sendError(ctx context.Context, roomUUID, userUUID, code, message string) error {
	metrics.MessageErrorsTotal.WithLabelValues(code).Inc()
	if c, ok := s.hub.Get(roomUUID, userUUID); ok {
		_ = c.Send(ctx, converter.ErrorEnvelope(roomUUID, code, message))
		metrics.MessagesTotal.WithLabelValues(string(model.TypeError), "out").Inc()
	}
	return nil
}
