package session

import (
	"context"
	"time"

	"signaling/internal/client/room"
	"signaling/internal/converter"
	"signaling/internal/metrics"
	"signaling/internal/model"
	"signaling/internal/repository"
	"signaling/internal/service"
	"signaling/internal/service/hub"
)

type Service struct {
	hub        *hub.Hub
	presence   repository.PresenceRepository
	roomClient room.RoomClient
	instanceID string
	roomCheck  bool
	sfuEnabled bool
	sfuURL     string
}

func New(
	h *hub.Hub,
	presence repository.PresenceRepository,
	roomClient room.RoomClient,
	instanceID string,
	roomCheck bool,
	sfuEnabled bool,
	sfuURL string,
) *Service {
	return &Service{
		hub:        h,
		presence:   presence,
		roomClient: roomClient,
		instanceID: instanceID,
		roomCheck:  roomCheck,
		sfuEnabled: sfuEnabled,
		sfuURL:     sfuURL,
	}
}

func (s *Service) Connect(ctx context.Context, conn service.Conn, roomUUID, userUUID string) error {
	if s.roomCheck && s.roomClient != nil {
		if err := s.roomClient.AssertCanJoin(ctx, roomUUID, userUUID); err != nil {
			return err
		}
	}

	evicted := s.hub.Register(roomUUID, userUUID, conn)
	if evicted != nil {
		_ = evicted.Send(ctx, converter.ErrorEnvelope(roomUUID, model.CodeSessionReplaced, "session replaced by new connection"))
		_ = evicted.Close(4000, "session replaced")
	}

	if err := s.presence.Register(ctx, roomUUID, userUUID, s.instanceID); err != nil {
		s.hub.Unregister(roomUUID, userUUID, conn)
		return err
	}
	_ = s.presence.EnsureSubscribed(ctx, roomUUID)

	peers, err := s.presence.ListPeers(ctx, roomUUID)
	if err != nil {
		peers = s.hub.LocalPeers(roomUUID)
	}
	otherPeers := make([]string, 0, len(peers))
	for _, p := range peers {
		if p != userUUID {
			otherPeers = append(otherPeers, p)
		}
	}

	welcome := model.Envelope{
		Type:     model.TypeWelcome,
		RoomUUID: roomUUID,
		Payload: converter.MustPayload(model.WelcomePayload{
			SelfUserUUID: userUUID,
			Peers:        otherPeers,
			SFU:          model.SFUHint{Enabled: s.sfuEnabled, URL: s.sfuURL},
		}),
		TS: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := conn.Send(ctx, welcome); err != nil {
		return err
	}
	metrics.MessagesTotal.WithLabelValues(string(model.TypeWelcome), "out").Inc()
	metrics.WSConnections.Inc()

	joined := model.Envelope{
		Type:         model.TypePeerJoined,
		RoomUUID:     roomUUID,
		FromUserUUID: userUUID,
		Payload:      converter.MustPayload(model.PeerEventPayload{UserUUID: userUUID}),
		TS:           time.Now().UTC().Format(time.RFC3339Nano),
	}
	s.hub.Broadcast(ctx, roomUUID, userUUID, joined)
	_ = s.presence.Publish(ctx, roomUUID, joined)
	metrics.MessagesTotal.WithLabelValues(string(model.TypePeerJoined), "out").Inc()

	return nil
}
