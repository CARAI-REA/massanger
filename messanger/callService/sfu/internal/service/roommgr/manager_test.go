package roommgr_test

import (
	"context"
	"encoding/json"
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

func sdpPayload(t *testing.T, msg model.Envelope) string {
	t.Helper()
	var payload model.SDPPayload
	require.NoError(t, json.Unmarshal(msg.Payload, &payload))
	require.NotEmpty(t, payload.SDP)
	return payload.SDP
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

// TestRenegotiateOfferAfterAddTrack covers the production path used after
// subscribePeerToTrack: AddTrack on a stable peer, then negotiate → TypeOffer.
func TestRenegotiateOfferAfterAddTrack(t *testing.T) {
	f, err := webrtcapi.NewFactory(webrtcapi.Config{})
	require.NoError(t, err)
	mgr := roommgr.NewManager(f, nil, roommgr.Options{})

	conn := &memConn{}
	peer, _, err := mgr.Join(context.Background(), "room", "user-a", conn)
	require.NoError(t, err)

	client, err := f.NewPeerConnection()
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close(); mgr.Leave("room", "user-a") })

	_, err = client.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
	})
	require.NoError(t, err)

	offer, err := client.CreateOffer(nil)
	require.NoError(t, err)
	require.NoError(t, client.SetLocalDescription(offer))
	offer = *client.LocalDescription()
	require.Contains(t, offer.SDP, "ice-ufrag")
	require.NoError(t, mgr.HandleSignal("room", "user-a", model.Envelope{
		Type:    model.TypeOffer,
		Payload: converter.MustPayload(model.SDPPayload{SDP: offer.SDP}),
	}))

	ansMsg := conn.waitType(t, model.TypeAnswer, 0, 2*time.Second)
	require.NoError(t, client.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdpPayload(t, ansMsg),
	}))

	// With no remote tracks to subscribe, the SFU must stay stable after the
	// initial answer (no speculative renegotiation offer).
	require.Equal(t, webrtc.SignalingStateStable, peer.PC.SignalingState())

	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio-late", "stream-late",
	)
	require.NoError(t, err)
	_, err = peer.PC.AddTrack(track)
	require.NoError(t, err)

	// Mirror onTrack → subscribePeerToTrack → negotiate.
	require.NoError(t, mgr.Negotiate("room", "user-a"))

	offerMsg := conn.waitType(t, model.TypeOffer, 0, 3*time.Second)
	require.Contains(t, sdpPayload(t, offerMsg), "audio-late")
}
