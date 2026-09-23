package session

import (
	"context"

	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
)

func (s *Service) CloseRoom(ctx context.Context, roomUUID string) error {
	errMsg := converter.ErrorEnvelope(roomUUID, model.CodeRoomClosed, "room closed")
	users := s.rooms.LocalUsers(roomUUID)
	for _, u := range users {
		if p, ok := s.rooms.Get(roomUUID, u); ok && p.Conn != nil {
			_ = p.Conn.SendAndClose(ctx, errMsg, 1000, "room closed")
			metrics.WSConnections.Dec()
		}
	}
	s.rooms.CloseRoom(roomUUID)
	_ = s.affinity.Release(ctx, roomUUID, s.instanceID)
	return nil
}

func (s *Service) CloseUser(ctx context.Context, roomUUID, userUUID string) error {
	p, ok := s.rooms.Get(roomUUID, userUUID)
	if !ok {
		return nil
	}
	_ = p.Conn.Send(ctx, converter.ErrorEnvelope(roomUUID, model.CodePermissionDenied, "removed from room"))
	return s.Disconnect(ctx, p.Conn, roomUUID, userUUID, "removed")
}
