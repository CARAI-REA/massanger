# callService WebRTC / Signaling protocol

## Roles

| Service | Port (internal) | Role |
|---------|-----------------|------|
| rooms | 8080 gRPC | Rooms, participants, join JWT, AssertCanJoin, RefreshJoinToken, TURN creds |
| signaling | 8081 WS `/v1/ws` | Presence, peer events; optional mesh SDP relay |
| sfu | 8082 WS `/v1/ws` | Media SFU (SDP/ICE + RTP fan-out) |

Production media path: **SFU**. Signaling `welcome.sfu` advertises the media URL when `SFU_ENABLED=true`.

TLS: terminate at **nginx** (`wss://`). Apps speak plain WS/HTTP on the private network.

## Auth

### User access JWT (rooms gRPC)

Header: `Authorization: Bearer <access_jwt>`  
Claim: `user_id` (UUID). Secret: `JWT_SECRET`.

### Join JWT (signaling + SFU WS)

- Production: **only** `Authorization: Bearer <join_token>` (`WS_ALLOW_QUERY_TOKEN=false`; process refuses to start if true).
- Development: Bearer or `?token=`.
- Query `room_uuid` required and must match JWT `room_uuid`.
- Secret: `JOIN_TOKEN_SECRET_KEY` (shared). Refresh via `RefreshJoinToken` (user access JWT) before expiry, then reconnect WS.

### Service JWT (signaling/sfu → rooms)

Header: `Authorization: Bearer <service_jwt>`  
Claim: `svc` ∈ `{signaling,sfu}`. Secret: `SERVICE_JWT_SECRET`.  
Used for: `AssertCanJoin`, `GetTURNCredentials` (SFU mints TURN into welcome — clients do **not** call GetTURNCredentials).

## Envelope (WS JSON)

```json
{
  "type": "welcome|peer_joined|peer_left|offer|answer|ice_candidate|ping|pong|bye|error",
  "room_uuid": "...",
  "from_user_uuid": "...",
  "to_user_uuid": "...",
  "payload": {},
  "ts": "RFC3339Nano"
}
```

### welcome (signaling)

- `self_user_uuid`, `peers[]`
- `sfu`: `{ "enabled": true, "url": "wss://calls.example/sfu/v1/ws" }` when configured

### welcome (sfu)

- `self_user_uuid`, `peers[]`, `role: "sfu"`
- `ice_servers[]`: STUN from config + **short-lived TURN** from Rooms `GetTURNCredentials` (HMAC username `expiry:userid`)

## SFU negotiation

1. Client connects WS (join JWT).
2. Client creates `RTCPeerConnection` with `ice_servers`, adds local tracks, sends **offer**.
3. SFU replies **answer** + ICE candidates.
4. When SFU adds remote tracks (late publisher / late joiner), SFU sends a renegotiation **offer**; client must **answer**.
5. Client ↔ SFU exchange `ice_candidate` until connected.

## Error codes (payload / close)

| Code | Close | Meaning |
|------|-------|---------|
| `TOKEN_EXPIRED` | **4001** | Join JWT expired — call `RefreshJoinToken`, reconnect |
| `INSTANCE_DRAINING` | **4002** | SFU draining; retry another instance / later |
| `ROOM_CLOSED` | — | Room ended; close PC |
| `SESSION_REPLACED` | 4000 | Same user connected elsewhere |
| `ROOM_ON_OTHER_INSTANCE` | — | Connect to `instance_url` in error payload |
| `PERMISSION_DENIED` | — | Not a participant |
| `TURN_UNAVAILABLE` | — | Prod SFU could not mint TURN (fail-closed) |

Expired join token: server upgrades WS briefly, sends `error` with `TOKEN_EXPIRED`, then closes with code **4001** (so clients can distinguish from invalid token HTTP 401).

## Multi-instance SFU

Redis affinity pins a room to one SFU instance. If wrong instance: error with `instance_url`. Client must reconnect to that URL (still via nginx if public URL is sticky-routed).

On SIGTERM / `POST /internal/drain` (localhost only): instance sets draining, `/readyz` fails, new Connects get `INSTANCE_DRAINING`.

## TURN

SFU fetches ephemeral credentials from Rooms on each Connect and puts them in `welcome.ice_servers`. Coturn: `use-auth-secret` + shared `TURN_SHARED_SECRET`.  
Set `WEBRTC_NAT_1TO1_IPS` to the public/LAN IP advertised for Docker/host ICE.

## Recording

Room setting `recording_enabled` is returned from `AssertCanJoin` and enables SFU raw RTP dump for that room (also `WEBRTC_RECORDING_ENABLED` global). HLS/S3 egress is **not** implemented.

## Sequence: two-party call

```mermaid
sequenceDiagram
  participant A as ClientA
  participant R as Rooms
  participant S as Signaling
  participant F as SFU
  A->>R: CreateRoom / AddParticipant
  R-->>A: join_token
  A->>S: WS join
  A->>F: WS join
  F->>R: AssertCanJoin + GetTURNCredentials
  F-->>A: welcome.ice_servers + answer path
  A->>F: offer
  F-->>A: answer + ice
  Note over A,F: peer B joins similarly; SFU may send renegotiation offer
  A->>R: EndRoom
  R-->>S: Kafka room_ended
```
