package session

import (
	"context"
	"time"

	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
	"sfu/internal/service"
)

func (s *Service) Disconnect(ctx context.Context, conn service.Conn, roomUUID, userUUID string, reason string) error {
	cur, ok := s.rooms.Get(roomUUID, userUUID)
	if !ok || (conn != nil && cur.Conn != conn) {
		if cur, ok := s.rooms.Get(roomUUID, userUUID); ok && cur.Conn != conn {
			metrics.WSConnections.Dec()
			if conn != nil {
				_ = conn.Close(1000, reason)
			}
			return nil
		}
		if conn != nil {
			_ = conn.Close(1000, reason)
		}
		return nil
	}

	s.rooms.Leave(roomUUID, userUUID)
	if len(s.rooms.LocalUsers(roomUUID)) == 0 {
		_ = s.affinity.Release(ctx, roomUUID, s.instanceID)
	}

	left := model.Envelope{
		Type:         model.TypePeerLeft,
		RoomUUID:     roomUUID,
		FromUserUUID: userUUID,
		Payload:      converter.MustPayload(model.PeerEventPayload{UserUUID: userUUID}),
		TS:           time.Now().UTC().Format(time.RFC3339Nano),
	}
	s.broadcast(ctx, roomUUID, userUUID, left)
	metrics.MessagesTotal.WithLabelValues(string(model.TypePeerLeft), "out").Inc()
	metrics.WSConnections.Dec()
	if conn != nil {
		_ = conn.Close(1000, reason)
	}
	return nil
}
