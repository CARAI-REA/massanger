# Signaling Service

WebSocket gateway for WebRTC signaling (SDP/ICE) and presence. Uses join JWT issued by Room Service.

## Ports

| Endpoint | Default |
|----------|---------|
| HTTP/WS  | `8081` (`/v1/ws`) |
| Metrics  | `9091` (`/metrics`) |
| Health   | `/healthz`, `/readyz` |

## Auth

Handshake only:

- Header `Authorization: Bearer <join_token>` (required in production)
- Query `token=` only when `WS_ALLOW_QUERY_TOKEN=true` (dev default)
- Query `room_uuid` required and must match JWT claim
- Secret: `JOIN_TOKEN_SECRET_KEY` (same as rooms)
- Rooms checks via `AssertCanJoin` with **service JWT** (`SERVICE_JWT_SECRET`)

See [docs/PROTOCOL.md](../docs/PROTOCOL.md) and [deploy/SECRETS.md](../deploy/SECRETS.md).

## Message types

| type | direction |
|------|-----------|
| welcome | out |
| peer_joined / peer_left | out |
| offer / answer / ice_candidate | in→out |
| ping / pong | both |
| bye | in |
| error | out |

Mesh P2P by default; when `SFU_ENABLED=true`, `welcome.sfu` advertises Media Server URL (`SFU_PUBLIC_URL`).

## Local run

```bash
# from callService/
task up-core
task up-rooms-postgres   # redis + kafka + rooms
cd signaling && go run ./cmd
```

Env file: `deploy/compose/signaling/.env`

```bash
task up-signaling        # full stack including signaling container
```

## Smoke

```bash
./scripts/smoke_signaling.sh
```

## Architecture

- `internal/api/ws/v1` — WebSocket upgrade + pumps
- `internal/service/hub` — in-memory sessions
- `internal/service/session` — connect/route/disconnect
- `internal/repository/presence` — Redis Set/Hash + Pub/Sub
- `internal/client/room` — gRPC AssertCanJoin
- `internal/kafka` — force-close on room_ended / participant_left
