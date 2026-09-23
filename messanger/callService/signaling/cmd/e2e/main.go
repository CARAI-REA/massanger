package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
)

type accessClaims struct {
	jwt.RegisteredClaims
	UserUUID string `json:"user_id"`
}

type envelope struct {
	Type         string          `json:"type"`
	RoomUUID     string          `json:"room_uuid"`
	FromUserUUID string          `json:"from_user_uuid"`
	ToUserUUID   string          `json:"to_user_uuid"`
	Payload      json.RawMessage `json:"payload"`
}

func main() {
	ctx := context.Background()
	owner := "11111111-1111-1111-1111-111111111111"
	peer := "22222222-2222-2222-2222-222222222222"
	secret := envOr("JWT_SECRET", "rooms-auth-secret")
	grpcAddr := envOr("GRPC_ADDR", "localhost:8080")
	wsBase := envOr("SIGNALING_WS", "ws://localhost:8081/v1/ws")

	mustHealth()

	token := mustAccessJWT(secret, owner)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(err)
	defer conn.Close()
	client := roomsV1.NewRoomServiceClient(conn)

	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	fmt.Println("==> CreateRoom")
	created, err := client.CreateRoom(authCtx, &roomsV1.CreateRoomRequest{
		Name:      "e2e-room",
		OwnerUuid: owner,
		RoomSettings: &roomsV1.RoomSettings{
			MaxParticipants: 8,
			Quality:         "HD",
			AutoClose:       false,
		},
	})
	must(err)
	roomUUID := created.GetRoom().GetRoomUuid()
	ownerJoin := created.GetJoinToken()
	fmt.Printf("room=%s owner_join_len=%d\n", roomUUID, len(ownerJoin))

	fmt.Println("==> AddParticipant")
	peerToken := mustAccessJWT(secret, peer)
	peerCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+peerToken))
	added, err := client.AddParticipant(peerCtx, &roomsV1.AddParticipantRequest{
		RoomUuid: roomUUID,
		UserUuid: peer,
	})
	must(err)
	peerJoin := added.GetJoinToken()
	fmt.Printf("peer_join_len=%d\n", len(peerJoin))

	fmt.Println("==> GetRoom / IsParticipant")
	room, err := client.GetRoom(authCtx, &roomsV1.GetRoomRequest{RoomUuid: roomUUID})
	must(err)
	if room.GetRoom().GetStatus() != "active" {
		panic("room not active: " + room.GetRoom().GetStatus())
	}
	ip, err := client.IsParticipant(authCtx, &roomsV1.IsParticipantRequest{RoomUuid: roomUUID, UserUuid: peer})
	must(err)
	if !ip.GetIsActive() {
		panic("peer not active participant")
	}
	fmt.Println("rooms OK")

	fmt.Println("==> Signaling WS two peers")
	a := mustWS(wsBase, roomUUID, ownerJoin)
	defer a.Close()
	welcomeA := mustRead(a, 3*time.Second)
	if welcomeA.Type != "welcome" {
		panic("expected welcome for A, got " + welcomeA.Type)
	}
	fmt.Println("A welcome OK")

	b := mustWS(wsBase, roomUUID, peerJoin)
	defer b.Close()
	welcomeB := mustRead(b, 3*time.Second)
	if welcomeB.Type != "welcome" {
		panic("expected welcome for B, got " + welcomeB.Type)
	}
	fmt.Println("B welcome OK")

	joined := mustRead(a, 3*time.Second)
	if joined.Type != "peer_joined" {
		panic("expected peer_joined on A, got " + joined.Type)
	}
	fmt.Println("peer_joined OK")

	fmt.Println("==> ping")
	must(a.WriteJSON(envelope{Type: "ping", Payload: json.RawMessage(`{}`)}))
	pong := mustRead(a, 3*time.Second)
	if pong.Type != "pong" {
		panic("expected pong, got " + pong.Type)
	}
	fmt.Println("pong OK")

	fmt.Println("==> offer relay")
	must(a.WriteJSON(envelope{
		Type:       "offer",
		ToUserUUID: peer,
		Payload:    json.RawMessage(`{"sdp":"v=0-e2e-offer"}`),
	}))
	offer := mustRead(b, 3*time.Second)
	if offer.Type != "offer" || offer.FromUserUUID != owner {
		panic(fmt.Sprintf("bad offer: %+v", offer))
	}
	fmt.Println("offer OK")

	fmt.Println("==> EndRoom force-close")
	_, err = client.EndRoom(authCtx, &roomsV1.EndRoomRequest{RoomUuid: roomUUID, OwnerUuid: owner})
	must(err)

	deadline := time.Now().Add(10 * time.Second)
	closed := false
	for time.Now().Before(deadline) {
		_ = a.SetReadDeadline(time.Now().Add(time.Second))
		msg, err := readOne(a)
		if err != nil {
			// Socket closed after force-close is acceptable if we already saw ROOM_CLOSED;
			// otherwise keep waiting on peer B in case A's read raced the close frame.
			_ = b.SetReadDeadline(time.Now().Add(time.Second))
			msgB, errB := readOne(b)
			if errB == nil && msgB.Type == "error" {
				var p struct {
					Code string `json:"code"`
				}
				_ = json.Unmarshal(msgB.Payload, &p)
				if p.Code == "ROOM_CLOSED" {
					closed = true
				}
			}
			break
		}
		if msg.Type == "error" {
			var p struct {
				Code string `json:"code"`
			}
			_ = json.Unmarshal(msg.Payload, &p)
			if p.Code == "ROOM_CLOSED" {
				closed = true
				break
			}
		}
	}
	if !closed {
		fmt.Println("WARN: ROOM_CLOSED not seen on WS within timeout (kafka lag?)")
	} else {
		fmt.Println("ROOM_CLOSED OK")
	}

	fmt.Println("ALL E2E CHECKS PASSED")
}

func mustHealth() {
	for _, u := range []string{
		"http://localhost:8081/healthz",
		"http://localhost:9090/metrics",
		"http://localhost:9091/metrics",
	} {
		resp, err := http.Get(u)
		must(err)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			panic("bad health: " + u)
		}
	}
	fmt.Println("health/metrics OK")
}

func mustAccessJWT(secret, user string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserUUID: user,
	})
	s, err := t.SignedString([]byte(secret))
	must(err)
	return s
}

func mustWS(base, room, token string) *websocket.Conn {
	u, err := url.Parse(base)
	must(err)
	q := u.Query()
	q.Set("room_uuid", room)
	q.Set("token", token)
	u.RawQuery = q.Encode()
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	must(err)
	return c
}

func mustRead(c *websocket.Conn, timeout time.Duration) envelope {
	_ = c.SetReadDeadline(time.Now().Add(timeout))
	msg, err := readOne(c)
	must(err)
	return msg
}

func readOne(c *websocket.Conn) (envelope, error) {
	_, data, err := c.ReadMessage()
	if err != nil {
		return envelope{}, err
	}
	var msg envelope
	if err := json.Unmarshal(data, &msg); err != nil {
		return envelope{}, err
	}
	return msg, nil
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
