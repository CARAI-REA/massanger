package session

import (
	"context"
	"time"

	"signaling/internal/converter"
	"signaling/internal/metrics"
	"signaling/internal/model"
	"signaling/internal/service"
)

func (s *Service) Disconnect(ctx context.Context, conn service.Conn, roomUUID, userUUID string, reason string) error {
	removed := s.hub.Unregister(roomUUID, userUUID, conn)
	if removed {
		_ = s.presence.Unregister(ctx, roomUUID, userUUID)
		_ = s.presence.ReleaseSubscribe(ctx, roomUUID)

		left := model.Envelope{
			Type:         model.TypePeerLeft,
			RoomUUID:     roomUUID,
			FromUserUUID: userUUID,
			Payload:      converter.MustPayload(model.PeerEventPayload{UserUUID: userUUID}),
			TS:           time.Now().UTC().Format(time.RFC3339Nano),
		}
		s.hub.Broadcast(ctx, roomUUID, userUUID, left)
		_ = s.presence.Publish(ctx, roomUUID, left)
		metrics.MessagesTotal.WithLabelValues(string(model.TypePeerLeft), "out").Inc()
		metrics.WSConnections.Dec()
		if conn != nil {
			_ = conn.Close(1000, reason)
		}
		return nil
	}

	// Replaced by a newer session for the same user.
	if cur, ok := s.hub.Get(roomUUID, userUUID); ok && cur != conn {
		metrics.WSConnections.Dec()
		if conn != nil {
			_ = conn.Close(1000, reason)
		}
		return nil
	}

	// Already removed (e.g. CloseRoom) — close socket only.
	if conn != nil {
		_ = conn.Close(1000, reason)
	}
	return nil
}

func (s *Service) OnRemoteMessage(ctx context.Context, roomUUID string, msg model.Envelope) {
	switch msg.Type {
	case model.TypePeerJoined, model.TypePeerLeft:
		// Originating instance already broadcast locally.
		if _, ok := s.hub.Get(roomUUID, msg.FromUserUUID); ok {
			return
		}
		s.hub.Broadcast(ctx, roomUUID, msg.FromUserUUID, msg)
	default:
		if msg.ToUserUUID == "" {
			s.hub.Broadcast(ctx, roomUUID, msg.FromUserUUID, msg)
			return
		}
		if c, ok := s.hub.Get(roomUUID, msg.ToUserUUID); ok {
			_ = c.Send(ctx, msg)
			metrics.RouteLocalTotal.Inc()
		}
	}
}
