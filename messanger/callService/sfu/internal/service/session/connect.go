package session

import (
	"context"
	"time"

	"sfu/internal/client/room"
	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
	"sfu/internal/repository"
	"sfu/internal/service"
	"sfu/internal/service/roommgr"
)

type Service struct {
	rooms      *roommgr.Manager
	affinity   repository.AffinityRepository
	roomClient room.RoomClient
	instanceID string
	publicURL  string
	affinityTTL time.Duration
	roomCheck  bool
	iceServers []model.ICEServer
}

func New(
	rooms *roommgr.Manager,
	affinity repository.AffinityRepository,
	roomClient room.RoomClient,
	instanceID, publicURL string,
	affinityTTL time.Duration,
	roomCheck bool,
) *Service {
	return &Service{
		rooms:       rooms,
		affinity:    affinity,
		roomClient:  roomClient,
		instanceID:  instanceID,
		publicURL:   publicURL,
		affinityTTL: affinityTTL,
		roomCheck:   roomCheck,
		iceServers:  rooms.ICEServers(),
	}
}

func (s *Service) Connect(ctx context.Context, conn service.Conn, roomUUID, userUUID string) error {
	if s.roomCheck && s.roomClient != nil {
		if err := s.roomClient.AssertCanJoin(ctx, roomUUID, userUUID); err != nil {
			return err
		}
	}

	claimed, ownerURL, err := s.affinity.Claim(ctx, roomUUID, s.instanceID, s.publicURL, s.affinityTTL)
	if err != nil {
		return err
	}
	if !claimed {
		metrics.AffinityRedirectsTotal.Inc()
		_ = conn.Send(ctx, converter.RedirectEnvelope(roomUUID, ownerURL))
		return model.ErrRoomOnOtherInstance
	}

	_, evicted, err := s.rooms.Join(ctx, roomUUID, userUUID, conn)
	if err != nil {
		_ = s.affinity.Release(ctx, roomUUID, s.instanceID)
		return err
	}
	if evicted != nil && evicted.Conn != nil {
		_ = evicted.Conn.SendAndClose(ctx,
			converter.ErrorEnvelope(roomUUID, model.CodeSessionReplaced, "session replaced"),
			4000, "session replaced")
	}

	peers := s.rooms.LocalUsers(roomUUID)
	other := make([]string, 0, len(peers))
	for _, p := range peers {
		if p != userUUID {
			other = append(other, p)
		}
	}

	welcome := model.Envelope{
		Type:     model.TypeWelcome,
		RoomUUID: roomUUID,
		Payload: converter.MustPayload(model.WelcomePayload{
			SelfUserUUID: userUUID,
			Peers:        other,
			ICEServers:   s.iceServers,
			Role:         "sfu",
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
	s.broadcast(ctx, roomUUID, userUUID, joined)
	metrics.MessagesTotal.WithLabelValues(string(model.TypePeerJoined), "out").Inc()
	return nil
}

func (s *Service) broadcast(ctx context.Context, roomUUID, exclude string, msg model.Envelope) {
	for _, u := range s.rooms.LocalUsers(roomUUID) {
		if u == exclude {
			continue
		}
		if p, ok := s.rooms.Get(roomUUID, u); ok && p.Conn != nil {
			_ = p.Conn.Send(ctx, msg)
		}
	}
}
