package session

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"sfu/internal/client/room"
	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
	"sfu/internal/repository"
	"sfu/internal/service"
	"sfu/internal/service/roommgr"
)

var tracer = otel.Tracer("sfu/session")

type Service struct {
	rooms              *roommgr.Manager
	affinity           repository.AffinityRepository
	roomClient         room.RoomClient
	instanceID         string
	publicURL          string
	affinityTTL        time.Duration
	roomCheck          bool
	requireEphemeralTURN bool
	iceServers         []model.ICEServer
	draining           atomic.Bool
}

func New(
	rooms *roommgr.Manager,
	affinity repository.AffinityRepository,
	roomClient room.RoomClient,
	instanceID, publicURL string,
	affinityTTL time.Duration,
	roomCheck bool,
	requireEphemeralTURN bool,
) *Service {
	return &Service{
		rooms:                rooms,
		affinity:             affinity,
		roomClient:           roomClient,
		instanceID:           instanceID,
		publicURL:            publicURL,
		affinityTTL:          affinityTTL,
		roomCheck:            roomCheck,
		requireEphemeralTURN: requireEphemeralTURN,
		iceServers:           rooms.ICEServers(),
	}
}

func (s *Service) BeginDrain() {
	s.draining.Store(true)
}

func (s *Service) IsDraining() bool {
	return s.draining.Load()
}

// WaitEmpty blocks until no local peers remain or ctx is done.
func (s *Service) WaitEmpty(ctx context.Context) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		if s.rooms.LocalPeerCount() == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Service) Connect(ctx context.Context, conn service.Conn, roomUUID, userUUID string) error {
	ctx, span := tracer.Start(ctx, "session.Connect")
	defer span.End()
	span.SetAttributes(
		attribute.String("room_uuid", roomUUID),
		attribute.String("user_uuid", userUUID),
	)

	if s.draining.Load() {
		_ = conn.Send(ctx, converter.ErrorEnvelope(roomUUID, model.CodeInstanceDraining, "instance draining"))
		return model.ErrInstanceDraining
	}

	recordingEnabled := false
	if s.roomCheck && s.roomClient != nil {
		check, err := s.roomClient.AssertCanJoin(ctx, roomUUID, userUUID)
		if err != nil {
			return err
		}
		recordingEnabled = check.RecordingEnabled
	}

	iceServers := s.iceServers
	if s.roomCheck && s.roomClient != nil {
		turnServers, err := s.roomClient.GetTURNCredentials(ctx, roomUUID, userUUID)
		if err != nil || len(turnServers) == 0 {
			if s.requireEphemeralTURN {
				if err == nil {
					err = model.ErrTURNUnavailable
				}
				return err
			}
		} else {
			iceServers = mergeICEServers(s.iceServers, turnServers)
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

	s.rooms.SetRoomRecording(roomUUID, recordingEnabled)

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
			ICEServers:   iceServers,
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

func mergeICEServers(staticServers, turnServers []model.ICEServer) []model.ICEServer {
	out := make([]model.ICEServer, 0, len(staticServers)+len(turnServers))
	for _, s := range staticServers {
		hasCred := s.Username != "" || s.Credential != ""
		if hasCred {
			continue // drop static TURN; ephemeral replaces it
		}
		out = append(out, s)
	}
	out = append(out, turnServers...)
	return out
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
