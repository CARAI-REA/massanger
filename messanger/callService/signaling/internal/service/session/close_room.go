package session

import (
	"context"

	"signaling/internal/converter"
	"signaling/internal/metrics"
	"signaling/internal/model"
)

func (s *Service) CloseRoom(ctx context.Context, roomUUID string) error {
	errMsg := converter.ErrorEnvelope(roomUUID, model.CodeRoomClosed, "room closed")
	peers := s.hub.LocalPeers(roomUUID)
	s.hub.CloseAll(ctx, roomUUID, errMsg)

	for _, userUUID := range peers {
		_ = s.presence.Unregister(ctx, roomUUID, userUUID)
		_ = s.presence.ReleaseSubscribe(ctx, roomUUID)
		metrics.WSConnections.Dec()
	}
	return nil
}

func (s *Service) CloseUser(ctx context.Context, roomUUID, userUUID string) error {
	c, ok := s.hub.Get(roomUUID, userUUID)
	if ok {
		_ = c.Send(ctx, converter.ErrorEnvelope(roomUUID, model.CodePermissionDenied, "removed from room"))
		return s.Disconnect(ctx, c, roomUUID, userUUID, "removed")
	}
	return nil
}
