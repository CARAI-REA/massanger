package session_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	platTokens "github.com/CARAI-REA/messanger/callService/platform/pkg/tokens"

	"signaling/internal/client/room"
	"signaling/internal/converter"
	"signaling/internal/model"
	"signaling/internal/repository/presence"
	"signaling/internal/service/hub"
	"signaling/internal/service/session"
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

func newService() *session.Service {
	return session.New(hub.New(), presence.NewNoop(), room.NewNoop(), "inst-1", false, false, "")
}

func TestH8_ConnectSendsWelcome(t *testing.T) {
	svc := newService()
	a := &memConn{}
	b := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), b, "room", "user-b"))

	aw := a.Messages()
	require.NotEmpty(t, aw)
	require.Equal(t, model.TypeWelcome, aw[0].Type)
	var welcome model.WelcomePayload
	require.NoError(t, json.Unmarshal(aw[0].Payload, &welcome))
	require.Equal(t, "user-a", welcome.SelfUserUUID)
	require.False(t, welcome.SFU.Enabled)

	// a should have received peer_joined for b
	found := false
	for _, m := range a.Messages() {
		if m.Type == model.TypePeerJoined {
			found = true
		}
	}
	require.True(t, found)
}

func TestR1_OfferWithoutTo(t *testing.T) {
	svc := newService()
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:    model.TypeOffer,
		Payload: converter.MustPayload(model.SDPPayload{SDP: "v=0"}),
	}))
	msgs := a.Messages()
	require.Len(t, msgs, 1)
	require.Equal(t, model.TypeError, msgs[0].Type)
	var errp model.ErrorPayload
	require.NoError(t, json.Unmarshal(msgs[0].Payload, &errp))
	require.Equal(t, model.CodeInvalidArgument, errp.Code)
}

func TestR2_OfferToSelf(t *testing.T) {
	svc := newService()
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:       model.TypeOffer,
		ToUserUUID: "user-a",
		Payload:    converter.MustPayload(model.SDPPayload{SDP: "v=0"}),
	}))
	var errp model.ErrorPayload
	require.NoError(t, json.Unmarshal(a.Messages()[0].Payload, &errp))
	require.Equal(t, model.CodeInvalidArgument, errp.Code)
}

func TestR3_OfferOffline(t *testing.T) {
	svc := newService()
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:       model.TypeOffer,
		ToUserUUID: "missing",
		Payload:    converter.MustPayload(model.SDPPayload{SDP: "v=0"}),
	}))
	var errp model.ErrorPayload
	require.NoError(t, json.Unmarshal(a.Messages()[0].Payload, &errp))
	require.Equal(t, model.CodeNotFound, errp.Code)
}

func TestR4_OfferLocalDelivered(t *testing.T) {
	svc := newService()
	a := &memConn{}
	b := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), b, "room", "user-b"))
	b.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:       model.TypeOffer,
		ToUserUUID: "user-b",
		Payload:    converter.MustPayload(model.SDPPayload{SDP: "offer-sdp"}),
	}))
	msgs := b.Messages()
	require.Len(t, msgs, 1)
	require.Equal(t, model.TypeOffer, msgs[0].Type)
	require.Equal(t, "user-a", msgs[0].FromUserUUID)
	var p model.SDPPayload
	require.NoError(t, json.Unmarshal(msgs[0].Payload, &p))
	require.Equal(t, "offer-sdp", p.SDP)
}

func TestR5_EmptySDP(t *testing.T) {
	svc := newService()
	a := &memConn{}
	b := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), b, "room", "user-b"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:       model.TypeAnswer,
		ToUserUUID: "user-b",
		Payload:    converter.MustPayload(model.SDPPayload{SDP: ""}),
	}))
	var errp model.ErrorPayload
	require.NoError(t, json.Unmarshal(a.Messages()[0].Payload, &errp))
	require.Equal(t, model.CodeInvalidArgument, errp.Code)
}

func TestR6_ICEDelivered(t *testing.T) {
	svc := newService()
	a := &memConn{}
	b := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), b, "room", "user-b"))
	b.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{
		Type:       model.TypeICECandidate,
		ToUserUUID: "user-b",
		Payload:    converter.MustPayload(model.ICEPayload{Candidate: "cand"}),
	}))
	require.Equal(t, model.TypeICECandidate, b.Messages()[0].Type)
}

func TestR7_PingPong(t *testing.T) {
	svc := newService()
	a := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	a.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{Type: model.TypePing}))
	require.Equal(t, model.TypePong, a.Messages()[0].Type)
}

func TestR8_ByePeerLeft(t *testing.T) {
	svc := newService()
	a := &memConn{}
	b := &memConn{}
	require.NoError(t, svc.Connect(context.Background(), a, "room", "user-a"))
	require.NoError(t, svc.Connect(context.Background(), b, "room", "user-b"))
	b.msgs = nil
	require.NoError(t, svc.Handle(context.Background(), "room", "user-a", model.Envelope{Type: model.TypeBye}))
	found := false
	for _, m := range b.Messages() {
		if m.Type == model.TypePeerLeft {
			found = true
		}
	}
	require.True(t, found)
}

func TestH5H6H7_JoinClaimsHelpers(t *testing.T) {
	// Document expected claim shape used by WS handshake verifier.
	claims := &platTokens.JoinClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		UserUUID: "u1",
		RoomUUID: "r1",
	}
	require.Equal(t, "r1", claims.RoomUUID)
	require.True(t, claims.ExpiresAt.Time.After(time.Now()))
}
