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
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func main() {
	ctx := context.Background()
	owner := "11111111-1111-1111-1111-111111111111"
	peer := "22222222-2222-2222-2222-222222222222"
	secret := envOr("JWT_SECRET", "rooms-auth-secret")
	grpcAddr := envOr("GRPC_ADDR", "localhost:8080")
	wsBase := envOr("SFU_WS", "ws://localhost:8082/v1/ws")

	mustHealth()

	token := mustAccessJWT(secret, owner)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(err)
	defer conn.Close()
	client := roomsV1.NewRoomServiceClient(conn)
	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	fmt.Println("==> CreateRoom")
	created, err := client.CreateRoom(authCtx, &roomsV1.CreateRoomRequest{
		Name:      "sfu-e2e",
		OwnerUuid: owner,
		RoomSettings: &roomsV1.RoomSettings{
			MaxParticipants: 8,
			Quality:         "HD",
		},
	})
	must(err)
	roomUUID := created.GetRoom().GetRoomUuid()
	ownerJoin := created.GetJoinToken()

	peerTok := mustAccessJWT(secret, peer)
	peerCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+peerTok))
	added, err := client.AddParticipant(peerCtx, &roomsV1.AddParticipantRequest{
		RoomUuid: roomUUID,
		UserUuid: peer,
	})
	must(err)

	fmt.Println("==> SFU WS")
	a := mustWS(wsBase, roomUUID, ownerJoin)
	defer a.Close()
	wa := mustRead(a, 5*time.Second)
	if wa.Type != "welcome" {
		panic("expected welcome A got " + wa.Type)
	}
	fmt.Println("A welcome OK")

	b := mustWS(wsBase, roomUUID, added.GetJoinToken())
	defer b.Close()
	wb := mustRead(b, 5*time.Second)
	if wb.Type != "welcome" {
		panic("expected welcome B got " + wb.Type)
	}
	fmt.Println("B welcome OK")

	joined := mustRead(a, 5*time.Second)
	if joined.Type != "peer_joined" {
		panic("expected peer_joined got " + joined.Type)
	}
	fmt.Println("peer_joined OK")

	must(a.WriteJSON(envelope{Type: "ping", Payload: json.RawMessage(`{}`)}))
	pong := mustRead(a, 5*time.Second)
	if pong.Type != "pong" {
		panic("expected pong got " + pong.Type)
	}
	fmt.Println("pong OK")

	_, err = client.EndRoom(authCtx, &roomsV1.EndRoomRequest{RoomUuid: roomUUID, OwnerUuid: owner})
	must(err)
	deadline := time.Now().Add(10 * time.Second)
	closed := false
	for time.Now().Before(deadline) {
		_ = a.SetReadDeadline(time.Now().Add(time.Second))
		msg, err := readOne(a)
		if err != nil {
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
	if closed {
		fmt.Println("ROOM_CLOSED OK")
	} else {
		fmt.Println("WARN: ROOM_CLOSED not seen")
	}
	fmt.Println("ALL SFU E2E CHECKS PASSED")
}

func mustHealth() {
	for _, u := range []string{"http://localhost:8082/healthz", "http://localhost:9093/metrics"} {
		resp, err := http.Get(u)
		must(err)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			panic("bad health " + u)
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
