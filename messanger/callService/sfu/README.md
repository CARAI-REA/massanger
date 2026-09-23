# SFU Media Server

Pion-based Selective Forwarding Unit. Clients exchange SDP/ICE with the SFU over WebSocket and publish/subscribe media tracks.

## Ports

| Endpoint | Default |
|----------|---------|
| HTTP/WS  | `8082` (`/v1/ws`) |
| Metrics  | `9093` (`/metrics`) |
| UDP ICE  | `10000-10100` |
| TURN     | `3478` (+ relay `49160-49200/udp`) via coturn in compose |

## Auth

Same join JWT as Signaling (`JOIN_TOKEN_SECRET_KEY` from Room Service).

- `Authorization: Bearer <join_token>` (prod: no `?token=`)
- Query `room_uuid` must match JWT claim
- Rooms `AssertCanJoin` via service JWT

Protocol: [docs/PROTOCOL.md](../docs/PROTOCOL.md)

## Control messages

welcome, offer, answer, ice_candidate, peer_joined, peer_left, ping/pong, bye, error

Media path is SFU (not mesh). Signaling remains for presence; set `SFU_ENABLED` / `SFU_PUBLIC_URL` on Signaling so `welcome.sfu` points here.

After `AddTrack` (late publisher / late joiner subscribe) the SFU sends a renegotiation **offer**; clients must answer it.

## ICE / TURN / Docker NAT

| Env | Purpose |
|-----|---------|
| `WEBRTC_ICE_SERVERS` | STUN/TURN URLs used by the SFU PeerConnection |
| `WEBRTC_CLIENT_ICE_SERVERS` | Optional override advertised in `welcome.ice_servers` (browsers) |
| `WEBRTC_TURN_USERNAME` / `WEBRTC_TURN_CREDENTIAL` | Applied to all `turn:` / `turns:` URLs |
| `WEBRTC_NAT_1TO1_IPS` | Public/LAN IPs for host candidates (fixes Docker bridge IPs) |
| `WEBRTC_NAT_1TO1_CANDIDATE_TYPE` | `host` (default) or `srflx` |

Local compose starts **coturn**. Set `WEBRTC_NAT_1TO1_IPS` to your machine LAN IP when clients are not on the same Docker host loopback. For internet, set a public IP / proper TURN `external-ip`.

## Local run

```bash
task up-core
task up-rooms-postgres
cd sfu && go run ./cmd
# or
task up-sfu
```

Env: `deploy/compose/sfu/.env`

## Smoke / E2E

```bash
./scripts/smoke_sfu.sh
cd sfu && go run ./cmd/e2e
```

## Architecture

- `internal/api/ws/v1` — WS control
- `internal/service/roommgr` — Pion PeerConnections + track fan-out + renegotiation
- `internal/service/session` — connect/route/disconnect
- `internal/repository/affinity` — Redis room→instance
- `internal/client/room` — gRPC AssertCanJoin
- `internal/kafka` — force-close
