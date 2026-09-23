package roommgr_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"

	"sfu/internal/converter"
	"sfu/internal/model"
	"sfu/internal/service/roommgr"
	webrtcapi "sfu/internal/webrtc"
)

type memConn struct {
	mu   sync.Mutex
	msgs []model.Envelope
}

func (c *memConn) Send(_ context.Context, msg model.Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, msg)
	return nil
}
func (c *memConn) SendAndClose(ctx context.Context, msg model.Envelope, _ int, _ string) error {
	return c.Send(ctx, msg)
}
func (c *memConn) Close(int, string) error { return nil }
func (c *memConn) RemoteAddr() string      { return "test" }

func (c *memConn) waitType(t *testing.T, typ model.MessageType, after int, timeout time.Duration) model.Envelope {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		n := 0
		for _, m := range c.msgs {
			if m.Type == typ {
				if n == after {
					c.mu.Unlock()
					return m
				}
				n++
			}
		}
		c.mu.Unlock()
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s (index %d)", typ, after)
	return model.Envelope{}
}

func TestRenegotiateOfferAfterAddTrack(t *testing.T) {
	f, err := webrtcapi.NewFactory(webrtcapi.Config{
		ICEServers: []webrtcapi.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	require.NoError(t, err)
	mgr := roommgr.NewManager(f, nil, roommgr.Options{})

	conn := &memConn{}
	peer, _, err := mgr.Join(context.Background(), "room", "user-a", conn)
	require.NoError(t, err)

	// Initial client offer → SFU answer (connection becomes stable).
	offerPC, err := f.NewPeerConnection()
	require.NoError(t, err)
	t.Cleanup(func() { _ = offerPC.Close(); mgr.Leave("room", "user-a") })

	_, err = offerPC.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
	})
	require.NoError(t, err)

	offer, err := offerPC.CreateOffer(nil)
	require.NoError(t, err)
	require.NoError(t, offerPC.SetLocalDescription(offer))
	offer = *offerPC.LocalDescription()
	require.Contains(t, offer.SDP, "ice-ufrag")
	require.NoError(t, mgr.HandleSignal("room", "user-a", model.Envelope{
		Type:    model.TypeOffer,
		Payload: converter.MustPayload(model.SDPPayload{SDP: offer.SDP}),
	}))
	_ = conn.waitType(t, model.TypeAnswer, 0, 2*time.Second)
	require.Equal(t, webrtc.SignalingStateStable, peer.PC.SignalingState())

	// Mid-call AddTrack must produce an SFU-originated renegotiation offer.
	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio-late", "stream-late",
	)
	require.NoError(t, err)
	_, err = peer.PC.AddTrack(track)
	require.NoError(t, err)

	_ = conn.waitType(t, model.TypeOffer, 0, 3*time.Second)
}

func TestFactoryNAT1To1(t *testing.T) {
	_, err := webrtcapi.NewFactory(webrtcapi.Config{
		ICEServers:           []webrtcapi.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
		NAT1To1IPs:           []string{"203.0.113.10"},
		NAT1To1CandidateType: "host",
		UDPPortMin:           10000,
		UDPPortMax:           10010,
	})
	require.NoError(t, err)

	_, err = webrtcapi.NewFactory(webrtcapi.Config{
		NAT1To1IPs:           []string{"203.0.113.10"},
		NAT1To1CandidateType: "bogus",
	})
	require.Error(t, err)
}
