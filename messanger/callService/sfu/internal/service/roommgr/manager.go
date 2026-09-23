package roommgr

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"

	"sfu/internal/converter"
	"sfu/internal/metrics"
	"sfu/internal/model"
	"sfu/internal/service"
	webrtcapi "sfu/internal/webrtc"
)

type Options struct {
	SimulcastEnabled bool
	RecordingEnabled bool
	RecordingDir     string
}

type publishedTrack struct {
	track  *webrtc.TrackRemote
	rid    string
	cancel context.CancelFunc
}

type Peer struct {
	UserUUID string
	Conn     service.Conn
	PC       *webrtc.PeerConnection

	mu                sync.Mutex
	negMu             sync.Mutex
	negotiationNeeded bool
	published         map[string]*publishedTrack // trackID -> remote
	subscribed        map[string]*webrtc.RTPSender
	cancel            context.CancelFunc
}

type Manager struct {
	factory    *webrtcapi.Factory
	iceServers []model.ICEServer
	opts       Options

	mu    sync.RWMutex
	rooms map[string]map[string]*Peer // room -> user -> peer
}

func NewManager(factory *webrtcapi.Factory, clientICE []webrtcapi.ICEServer, opts Options) *Manager {
	servers := make([]model.ICEServer, 0, len(clientICE))
	for _, s := range clientICE {
		servers = append(servers, model.ICEServer{
			URLs:       append([]string(nil), s.URLs...),
			Username:   s.Username,
			Credential: s.Credential,
		})
	}
	if len(servers) == 0 {
		for _, s := range factory.ICEServers() {
			servers = append(servers, model.ICEServer{
				URLs:       append([]string(nil), s.URLs...),
				Username:   s.Username,
				Credential: s.Credential,
			})
		}
	}
	if opts.RecordingDir == "" {
		opts.RecordingDir = "./recordings"
	}
	return &Manager{
		factory:    factory,
		iceServers: servers,
		opts:       opts,
		rooms:      make(map[string]map[string]*Peer),
	}
}

func (m *Manager) ICEServers() []model.ICEServer { return m.iceServers }

func (m *Manager) RoomUUIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.rooms))
	for roomUUID := range m.rooms {
		out = append(out, roomUUID)
	}
	return out
}

func (m *Manager) LocalUsers(roomUUID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	peers := m.rooms[roomUUID]
	out := make([]string, 0, len(peers))
	for u := range peers {
		out = append(out, u)
	}
	return out
}

func (m *Manager) Get(roomUUID, userUUID string) (*Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.rooms[roomUUID][userUUID]
	return p, ok
}

func (m *Manager) Join(ctx context.Context, roomUUID, userUUID string, conn service.Conn) (*Peer, *Peer, error) {
	pc, err := m.factory.NewPeerConnection()
	if err != nil {
		return nil, nil, err
	}

	peerCtx, cancel := context.WithCancel(context.Background())
	p := &Peer{
		UserUUID:   userUUID,
		Conn:       conn,
		PC:         pc,
		published:  make(map[string]*publishedTrack),
		subscribed: make(map[string]*webrtc.RTPSender),
		cancel:     cancel,
	}

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		cand := c.ToJSON()
		var idx *uint16
		if cand.SDPMLineIndex != nil {
			v := *cand.SDPMLineIndex
			idx = &v
		}
		mid := ""
		if cand.SDPMid != nil {
			mid = *cand.SDPMid
		}
		_ = conn.Send(context.Background(), model.Envelope{
			Type:         model.TypeICECandidate,
			RoomUUID:     roomUUID,
			FromUserUUID: "",
			Payload: converter.MustPayload(model.ICEPayload{
				Candidate:     cand.Candidate,
				SDPMID:        mid,
				SDPMLineIndex: idx,
			}),
			TS: time.Now().UTC().Format(time.RFC3339Nano),
		})
		metrics.MessagesTotal.WithLabelValues(string(model.TypeICECandidate), "out").Inc()
	})

	pc.OnTrack(func(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		m.onTrack(peerCtx, roomUUID, userUUID, remote)
	})

	pc.OnNegotiationNeeded(func() {
		go m.negotiate(roomUUID, p)
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			cancel()
		}
	})

	m.mu.Lock()
	if m.rooms[roomUUID] == nil {
		m.rooms[roomUUID] = make(map[string]*Peer)
	}
	var evicted *Peer
	if old, ok := m.rooms[roomUUID][userUUID]; ok && old != p {
		evicted = old
	}
	m.rooms[roomUUID][userUUID] = p
	m.mu.Unlock()

	if evicted != nil {
		evicted.shutdown()
	}

	// Subscribe new peer to existing published tracks.
	m.mu.RLock()
	others := make([]*Peer, 0)
	for uid, other := range m.rooms[roomUUID] {
		if uid == userUUID {
			continue
		}
		others = append(others, other)
	}
	m.mu.RUnlock()
	subscribed := false
	for _, other := range others {
		other.mu.Lock()
		for _, pub := range other.published {
			if err := m.subscribePeerToTrack(p, pub.track); err == nil {
				subscribed = true
			}
		}
		other.mu.Unlock()
	}
	// Only mark renegotiation when the new peer actually got remote tracks.
	// Otherwise post-answer negotiate would race into HaveLocalOffer for nothing.
	if subscribed {
		m.negotiate(roomUUID, p)
	}

	metrics.PeerConnections.Inc()
	_ = ctx
	return p, evicted, nil
}

func (m *Manager) shouldSkipSimulcastRID(pub *Peer, remote *webrtc.TrackRemote) bool {
	if !m.opts.SimulcastEnabled {
		return false
	}
	rid := remote.RID()
	if rid == "" {
		return false
	}
	log.Printf("sfu simulcast: track id=%s stream=%s rid=%s", remote.ID(), remote.StreamID(), rid)
	if rid != "q" {
		return false
	}
	streamID := remote.StreamID()
	pub.mu.Lock()
	defer pub.mu.Unlock()
	for _, pt := range pub.published {
		if pt.track == nil || pt.track.StreamID() != streamID {
			continue
		}
		switch pt.rid {
		case "h", "m", "":
			return true
		}
	}
	return false
}

func (m *Manager) onTrack(ctx context.Context, roomUUID, publisherUUID string, remote *webrtc.TrackRemote) {
	m.mu.RLock()
	pub, ok := m.rooms[roomUUID][publisherUUID]
	m.mu.RUnlock()
	if !ok {
		return
	}

	if m.shouldSkipSimulcastRID(pub, remote) {
		log.Printf("sfu simulcast: skipping low rid=q for stream=%s", remote.StreamID())
		return
	}

	trackCtx, trackCancel := context.WithCancel(ctx)
	rid := remote.RID()
	pub.mu.Lock()
	pub.published[remote.ID()] = &publishedTrack{track: remote, rid: rid, cancel: trackCancel}
	pub.mu.Unlock()
	metrics.TracksPublished.Inc()

	m.mu.RLock()
	subscribers := make([]*Peer, 0)
	for uid, peer := range m.rooms[roomUUID] {
		if uid == publisherUUID {
			continue
		}
		subscribers = append(subscribers, peer)
	}
	m.mu.RUnlock()

	for _, sub := range subscribers {
		if err := m.subscribePeerToTrack(sub, remote); err == nil {
			m.negotiate(roomUUID, sub)
		}
	}

	go func() {
		var recFile *os.File
		if m.opts.RecordingEnabled {
			recFile = m.openRecordingFile(roomUUID, remote)
		}
		defer func() {
			if recFile != nil {
				_ = recFile.Close()
			}
			trackCancel()
			pub.mu.Lock()
			delete(pub.published, remote.ID())
			pub.mu.Unlock()
			metrics.TracksPublished.Dec()
		}()
		for {
			select {
			case <-trackCtx.Done():
				return
			default:
			}
			pkt, _, err := remote.ReadRTP()
			if err != nil {
				return
			}
			metrics.RTPPacketsTotal.WithLabelValues("in").Inc()
			if recFile != nil {
				if raw, mErr := pkt.Marshal(); mErr == nil {
					_, _ = recFile.Write(raw)
				}
			}
			m.mu.RLock()
			peers := m.rooms[roomUUID]
			m.mu.RUnlock()
			for uid, peer := range peers {
				if uid == publisherUUID {
					continue
				}
				peer.mu.Lock()
				sender, ok := peer.subscribed[remote.ID()]
				peer.mu.Unlock()
				if !ok || sender == nil || sender.Track() == nil {
					continue
				}
				if local, ok := sender.Track().(*webrtc.TrackLocalStaticRTP); ok {
					if err := local.WriteRTP(pkt); err == nil {
						metrics.RTPPacketsTotal.WithLabelValues("out").Inc()
					}
				}
			}
		}
	}()

	// PLI for video keyframes
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-trackCtx.Done():
				return
			case <-ticker.C:
				m.mu.RLock()
				pubPeer := m.rooms[roomUUID][publisherUUID]
				m.mu.RUnlock()
				if pubPeer == nil || pubPeer.PC == nil {
					return
				}
				_ = pubPeer.PC.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{MediaSSRC: uint32(remote.SSRC())},
				})
			}
		}
	}()
}

func (m *Manager) openRecordingFile(roomUUID string, remote *webrtc.TrackRemote) *os.File {
	dir := filepath.Join(m.opts.RecordingDir, sanitizePathPart(roomUUID))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		log.Printf("sfu recording: mkdir %s: %v", dir, err)
		return nil
	}
	name := sanitizePathPart(remote.ID())
	if name == "" {
		name = "track"
	}
	if rid := remote.RID(); rid != "" {
		name = name + "_" + sanitizePathPart(rid)
	}
	path := filepath.Join(dir, name+".rtp")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		log.Printf("sfu recording: open %s: %v", path, err)
		return nil
	}
	log.Printf("sfu recording: writing %s", path)
	return f
}

func sanitizePathPart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func (m *Manager) subscribePeerToTrack(sub *Peer, remote *webrtc.TrackRemote) error {
	if m.opts.SimulcastEnabled {
		rid := remote.RID()
		if rid == "h" || rid == "m" || rid == "" {
			log.Printf("sfu simulcast: prefer subscribe rid=%q track=%s", rid, remote.ID())
		}
	}

	sub.mu.Lock()
	if _, exists := sub.subscribed[remote.ID()]; exists {
		sub.mu.Unlock()
		return nil
	}
	sub.mu.Unlock()

	local, err := webrtc.NewTrackLocalStaticRTP(remote.Codec().RTPCodecCapability, remote.ID(), remote.StreamID())
	if err != nil {
		return err
	}
	sender, err := sub.PC.AddTrack(local)
	if err != nil {
		return err
	}
	sub.mu.Lock()
	sub.subscribed[remote.ID()] = sender
	sub.mu.Unlock()
	metrics.TracksSubscribed.Inc()

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	return nil
}

// negotiate pushes an SFU-originated offer when tracks were added after the peer is stable.
// Before the client's first offer, we only mark negotiationNeeded; CreateAnswer / a follow-up
// offer after SetRemoteDescription(answer) will pick it up.
func (m *Manager) negotiate(roomUUID string, p *Peer) {
	if p == nil || p.PC == nil {
		return
	}
	p.negMu.Lock()
	defer p.negMu.Unlock()

	if p.PC.RemoteDescription() == nil {
		p.negotiationNeeded = true
		return
	}
	if p.PC.SignalingState() != webrtc.SignalingStateStable {
		p.negotiationNeeded = true
		return
	}

	offer, err := p.PC.CreateOffer(nil)
	if err != nil {
		return
	}
	if err := p.PC.SetLocalDescription(offer); err != nil {
		return
	}
	p.negotiationNeeded = false
	_ = p.Conn.Send(context.Background(), model.Envelope{
		Type:     model.TypeOffer,
		RoomUUID: roomUUID,
		Payload:  converter.MustPayload(model.SDPPayload{SDP: offer.SDP}),
		TS:       time.Now().UTC().Format(time.RFC3339Nano),
	})
	metrics.MessagesTotal.WithLabelValues(string(model.TypeOffer), "out").Inc()
}

func (m *Manager) HandleSignal(roomUUID, userUUID string, msg model.Envelope) error {
	p, ok := m.Get(roomUUID, userUUID)
	if !ok || p.PC == nil {
		return model.ErrNotFound
	}
	switch msg.Type {
	case model.TypeOffer:
		var payload model.SDPPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.SDP == "" {
			return model.ErrInvalidArgument
		}
		if err := p.PC.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  payload.SDP,
		}); err != nil {
			return err
		}
		answer, err := p.PC.CreateAnswer(nil)
		if err != nil {
			return err
		}
		if err := p.PC.SetLocalDescription(answer); err != nil {
			return err
		}
		if err := p.Conn.Send(context.Background(), model.Envelope{
			Type:     model.TypeAnswer,
			RoomUUID: roomUUID,
			Payload:  converter.MustPayload(model.SDPPayload{SDP: answer.SDP}),
			TS:       time.Now().UTC().Format(time.RFC3339Nano),
		}); err != nil {
			return err
		}
		metrics.MessagesTotal.WithLabelValues(string(model.TypeAnswer), "out").Inc()
		// After initial answer, push SFU offer only if subscribers still need m-lines.
		p.negMu.Lock()
		needed := p.negotiationNeeded
		p.negMu.Unlock()
		if needed {
			go m.negotiate(roomUUID, p)
		}
		return nil
	case model.TypeAnswer:
		var payload model.SDPPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.SDP == "" {
			return model.ErrInvalidArgument
		}
		if err := p.PC.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  payload.SDP,
		}); err != nil {
			return err
		}
		p.negMu.Lock()
		needed := p.negotiationNeeded
		p.negMu.Unlock()
		if needed {
			go m.negotiate(roomUUID, p)
		}
		return nil
	case model.TypeICECandidate:
		var payload model.ICEPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.Candidate == "" {
			return model.ErrInvalidArgument
		}
		cand := webrtc.ICECandidateInit{Candidate: payload.Candidate}
		if payload.SDPMID != "" {
			cand.SDPMid = &payload.SDPMID
		}
		if payload.SDPMLineIndex != nil {
			cand.SDPMLineIndex = payload.SDPMLineIndex
		}
		return p.PC.AddICECandidate(cand)
	default:
		return model.ErrInvalidArgument
	}
}

// Negotiate pushes an SFU-originated offer for a peer (same path as after
// subscribePeerToTrack when a new remote track is published).
func (m *Manager) Negotiate(roomUUID, userUUID string) error {
	p, ok := m.Get(roomUUID, userUUID)
	if !ok || p.PC == nil {
		return model.ErrNotFound
	}
	m.negotiate(roomUUID, p)
	return nil
}

func (m *Manager) Leave(roomUUID, userUUID string) {
	m.mu.Lock()
	peers := m.rooms[roomUUID]
	if peers == nil {
		m.mu.Unlock()
		return
	}
	p, ok := peers[userUUID]
	if ok {
		delete(peers, userUUID)
	}
	if len(peers) == 0 {
		delete(m.rooms, roomUUID)
	}
	m.mu.Unlock()
	if ok {
		p.shutdown()
		metrics.PeerConnections.Dec()
	}
}

func (m *Manager) CloseRoom(roomUUID string) {
	m.mu.Lock()
	peers := m.rooms[roomUUID]
	delete(m.rooms, roomUUID)
	m.mu.Unlock()
	for _, p := range peers {
		p.shutdown()
		metrics.PeerConnections.Dec()
	}
}

func (m *Manager) CloseAll() {
	m.mu.Lock()
	rooms := m.rooms
	m.rooms = make(map[string]map[string]*Peer)
	m.mu.Unlock()
	for _, peers := range rooms {
		for _, p := range peers {
			p.shutdown()
			metrics.PeerConnections.Dec()
		}
	}
}

func (p *Peer) shutdown() {
	if p.cancel != nil {
		p.cancel()
	}
	p.mu.Lock()
	for _, pub := range p.published {
		if pub.cancel != nil {
			pub.cancel()
		}
	}
	p.published = make(map[string]*publishedTrack)
	for range p.subscribed {
		metrics.TracksSubscribed.Dec()
	}
	p.subscribed = make(map[string]*webrtc.RTPSender)
	p.mu.Unlock()
	if p.PC != nil {
		_ = p.PC.Close()
	}
}
