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
	sigWS := envOr("SIGNALING_WS", "ws://localhost:8081/v1/ws")
	sfuWS := envOr("SFU_WS", "ws://localhost:8082/v1/ws")

	step("1. Health checks (rooms / signaling / sfu)")
	mustHealth("http://localhost:9090/metrics")
	mustHTTP("http://localhost:8081/healthz")
	mustHTTP("http://localhost:8081/readyz")
	mustHTTP("http://localhost:8082/healthz")
	mustHTTP("http://localhost:8082/readyz")
	fmt.Println("OK health")

	step("2. Rooms: CreateRoom + AddParticipant")
	token := mustAccessJWT(secret, owner)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	must(err)
	defer conn.Close()
	client := roomsV1.NewRoomServiceClient(conn)
	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	created, err := client.CreateRoom(authCtx, &roomsV1.CreateRoomRequest{
		Name:      "full-stack-e2e",
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
	fmt.Printf("room=%s\n", roomUUID)

	peerTok := mustAccessJWT(secret, peer)
	peerCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+peerTok))
	added, err := client.AddParticipant(peerCtx, &roomsV1.AddParticipantRequest{
		RoomUuid: roomUUID,
		UserUuid: peer,
	})
	must(err)
	peerJoin := added.GetJoinToken()

	room, err := client.GetRoom(authCtx, &roomsV1.GetRoomRequest{RoomUuid: roomUUID})
	must(err)
	if room.GetRoom().GetStatus() != "active" {
		panic("room not active")
	}
	ip, err := client.IsParticipant(authCtx, &roomsV1.IsParticipantRequest{RoomUuid: roomUUID, UserUuid: peer})
	must(err)
	if !ip.GetIsActive() {
		panic("peer not participant")
	}
	fmt.Println("OK rooms")

	step("3. Signaling: welcome.sfu + peer_joined + ping")
	sa := mustWS(sigWS, roomUUID, ownerJoin)
	defer sa.Close()
	sw := mustRead(sa, 5*time.Second)
	if sw.Type != "welcome" {
		panic("signaling expected welcome, got " + sw.Type)
	}
	var welcome struct {
		SFU struct {
			Enabled bool   `json:"enabled"`
			URL     string `json:"url"`
		} `json:"sfu"`
	}
	must(json.Unmarshal(sw.Payload, &welcome))
	if !welcome.SFU.Enabled {
		panic("signaling welcome.sfu.enabled=false")
	}
	if welcome.SFU.URL == "" {
		panic("signaling welcome.sfu.url empty")
	}
	fmt.Printf("signaling sfu.url=%s\n", welcome.SFU.URL)

	sb := mustWS(sigWS, roomUUID, peerJoin)
	defer sb.Close()
	if mustRead(sb, 5*time.Second).Type != "welcome" {
		panic("signaling B welcome")
	}
	if mustRead(sa, 5*time.Second).Type != "peer_joined" {
		panic("signaling peer_joined")
	}
	must(sa.WriteJSON(envelope{Type: "ping", Payload: json.RawMessage(`{}`)}))
	if mustRead(sa, 5*time.Second).Type != "pong" {
		panic("signaling pong")
	}
	fmt.Println("OK signaling")

	step("4. SFU: welcome + peer_joined + ping")
	fa := mustWS(sfuWS, roomUUID, ownerJoin)
	defer fa.Close()
	fw := mustRead(fa, 5*time.Second)
	if fw.Type != "welcome" {
		panic("sfu expected welcome, got " + fw.Type)
	}
	var sfuWelcome struct {
		Role string `json:"role"`
	}
	must(json.Unmarshal(fw.Payload, &sfuWelcome))
	if sfuWelcome.Role != "sfu" {
		panic("sfu welcome.role != sfu")
	}

	fb := mustWS(sfuWS, roomUUID, peerJoin)
	defer fb.Close()
	if mustRead(fb, 5*time.Second).Type != "welcome" {
		panic("sfu B welcome")
	}
	if mustRead(fa, 5*time.Second).Type != "peer_joined" {
		panic("sfu peer_joined")
	}
	must(fa.WriteJSON(envelope{Type: "ping", Payload: json.RawMessage(`{}`)}))
	if mustRead(fa, 5*time.Second).Type != "pong" {
		panic("sfu pong")
	}
	fmt.Println("OK sfu")

	step("5. EndRoom → ROOM_CLOSED on Signaling + SFU")
	_, err = client.EndRoom(authCtx, &roomsV1.EndRoomRequest{RoomUuid: roomUUID, OwnerUuid: owner})
	must(err)

	sigClosed := waitRoomClosed(sa, 12*time.Second)
	sfuClosed := waitRoomClosed(fa, 12*time.Second)
	if !sigClosed {
		panic("signaling did not get ROOM_CLOSED")
	}
	if !sfuClosed {
		panic("sfu did not get ROOM_CLOSED")
	}
	fmt.Println("OK force-close")

	fmt.Println("\n=== FULL STACK E2E PASSED (rooms + signaling + sfu) ===")
}

func step(s string) { fmt.Println("\n==>", s) }

func waitRoomClosed(c *websocket.Conn, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_ = c.SetReadDeadline(time.Now().Add(time.Second))
		msg, err := readOne(c)
		if err != nil {
			return false
		}
		if msg.Type == "error" {
			var p struct {
				Code string `json:"code"`
			}
			_ = json.Unmarshal(msg.Payload, &p)
			if p.Code == "ROOM_CLOSED" {
				return true
			}
		}
	}
	return false
}

func mustHealth(u string) {
	resp, err := http.Get(u)
	must(err)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("bad " + u)
	}
}

func mustHTTP(u string) {
	resp, err := http.Get(u)
	must(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("bad " + u)
	}
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
