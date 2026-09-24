package session_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"sfu/internal/client/room"
	"sfu/internal/converter"
	"sfu/internal/model"
	"sfu/internal/repository/affinity"
	"sfu/internal/service/roommgr"
	"sfu/internal/service/session"
	webrtcapi "sfu/internal/webrtc"
)

type memConn struct {
	mu     sync.Mutex
	msgs   []model.Envelope
	closed bool
}

func (c *memConn) Send(_ context.Context, msg model.Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, msg)
	return nil
}

func (c *memConn) SendAndClose(ctx context.Context, msg model.Envelope, code int, reason string) error {
	_ = c.Send(ctx, msg)
	return c.Close(code, reason)
}

func (c *memConn) Close(int, string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *memConn) RemoteAddr() string { return "test" }

func (c *memConn) Messages() []model.Envelope {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]model.Envelope, len(c.msgs))
	copy(out, c.msgs)
	return out
}

func newSvc(t *testing.T) *session.Service {
	t.Helper()
	f, err := webrtcapi.NewFactory(webrtcapi.Config{
		ICEServers: []webrtcapi.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	require.NoError(t, err)
	mgr := roommgr.NewManager(f, nil, roommgr.Options{})
	return session.New(mgr, affinity.NewNoop(), room.NewNoop(), "inst-1", "ws://localhost:8082/v1/ws", 30*time.Second, false, false)
}

func TestConnectWelcome(t *testing.T) {
	svc := newSvc(t)
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	msgs := a.Messages()
	require.NotEmpty(t, msgs)
	require.Equal(t, model.TypeWelcome, msgs[0].Type)
	var w model.WelcomePayload
	require.NoError(t, json.Unmarshal(msgs[0].Payload, &w))
	require.Equal(t, "user-a", w.SelfUserUUID)
	require.Equal(t, "sfu", w.Role)
}

type turnMockClient struct {
	servers []model.ICEServer
}

func (m *turnMockClient) AssertCanJoin(context.Context, string, string) (room.JoinCheck, error) {
	return room.JoinCheck{OK: true, RecordingEnabled: true}, nil
}

func (m *turnMockClient) GetTURNCredentials(context.Context, string, string) ([]model.ICEServer, error) {
	return m.servers, nil
}

func TestConnectWelcomeEphemeralTURN(t *testing.T) {
	f, err := webrtcapi.NewFactory(webrtcapi.Config{
		ICEServers: []webrtcapi.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	require.NoError(t, err)
	mgr := roommgr.NewManager(f, []webrtcapi.ICEServer{
		{URLs: []string{"stun:stun.l.google.com:19302"}},
		{URLs: []string{"turn:turn.example:3478"}, Username: "static", Credential: "static"},
	}, roommgr.Options{})
	mock := &turnMockClient{servers: []model.ICEServer{{
		URLs:       []string{"turn:turn.example:3478?transport=udp"},
		Username:   "12345:user-a",
		Credential: "ephemeral-hmac",
	}}}
	svc := session.New(mgr, affinity.NewNoop(), mock, "inst-1", "ws://localhost:8082/v1/ws", 30*time.Second, true, true)
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	var w model.WelcomePayload
	require.NoError(t, json.Unmarshal(a.Messages()[0].Payload, &w))
	require.NotEmpty(t, w.ICEServers)
	found := false
	for _, s := range w.ICEServers {
		if s.Username == "12345:user-a" && s.Credential == "ephemeral-hmac" {
			found = true
		}
		require.NotEqual(t, "static", s.Username)
	}
	require.True(t, found)
}

func TestConnectRejectsWhenDraining(t *testing.T) {
	svc := newSvc(t)
	svc.BeginDrain()
	a := &memConn{}
	require.ErrorIs(t, svc.Connect(context.Background(), a, "room", "user-a"), model.ErrInstanceDraining)
}

func TestPeerJoined(t *testing.T) {
	svc := newSvc(t)
	a, b := &memConn{}, &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), b, "room", "user-b"))
	found := false
	for _, m := range a.Messages() {
		if m.Type == model.TypePeerJoined {
			found = true
		}
	}
	require.True(t, found)
}

func TestPingPong(t *testing.T) {
	svc := newSvc(t)
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{Type: model.TypePing}))
	require.Equal(t, model.TypePong, a.Messages()[0].Type)
}

func TestOfferEmptySDP(t *testing.T) {
	svc := newSvc(t)
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:    model.TypeOffer,
		Payload: converter.MustPayload(model.SDPPayload{SDP: ""}),
	}))
	var errp model.ErrorPayload
	require.NoError(t, json.Unmarshal(a.Messages()[0].Payload, &errp))
	require.Equal(t, model.CodeInvalidArgument, errp.Code)
}

func TestSessionReplaced(t *testing.T) {
	svc := newSvc(t)
	a, a2 := &memConn{}, &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), a2, "room", "user-a"))
	require.True(t, a.closed)
}
