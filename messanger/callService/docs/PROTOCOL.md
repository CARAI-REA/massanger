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

- Production: **only** `Authorization: Bearer <join_token>` (`WS_ALLOW_QUERY_TOKEN=false`).
- Development: Bearer or `?token=`.
- Query `room_uuid` required and must match JWT `room_uuid`.
- Secret: `JOIN_TOKEN_SECRET_KEY` (shared). Refresh via `RefreshJoinToken` before expiry.

### Service JWT (signaling/sfu → rooms)

Header: `Authorization: Bearer <service_jwt>`  
Claim: `svc` ∈ `{signaling,sfu}`. Secret: `SERVICE_JWT_SECRET`.  
Used for: `AssertCanJoin`, `GetTURNCredentials` (service path).

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
- `ice_servers[]`: `{ "urls": ["stun:...","turn:..."], "username?", "credential?" }`

## SFU negotiation

1. Client connects WS (join JWT).
2. Client creates `RTCPeerConnection` with `ice_servers`, adds local tracks, sends **offer**.
3. SFU replies **answer** + ICE candidates.
4. When SFU adds remote tracks (late publisher / late joiner), SFU sends a renegotiation **offer**; client must **answer**.
5. Client ↔ SFU exchange `ice_candidate` until connected.

## Error codes (payload / close)

| Code | Meaning |
|------|---------|
| `TOKEN_EXPIRED` / unauthenticated | Refresh join token and reconnect |
| `ROOM_CLOSED` | Room ended; close PC |
| `SESSION_REPLACED` | Same user connected elsewhere |
| `ROOM_ON_OTHER_INSTANCE` | Connect to `instance_url` in error payload |
| `PERMISSION_DENIED` | Not a participant |

## Multi-instance SFU

Redis affinity pins a room to one SFU instance. If wrong instance: error with `instance_url`. Client must reconnect to that URL (still via nginx if public URL is sticky-routed).

## TURN

Prefer short-lived HMAC credentials from Rooms `GetTURNCredentials` (coturn `use-auth-secret`).  
Set `WEBRTC_NAT_1TO1_IPS` to the public/LAN IP advertised for Docker/host ICE.

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
  A->>F: WS join + offer
  F-->>A: answer + ice
  Note over A,F: peer B joins similarly; SFU may send renegotiation offer
  A->>R: EndRoom
  R-->>S: Kafka room_ended
  R-->>F: Kafka room_ended
  S-->>A: ROOM_CLOSED
  F-->>A: ROOM_CLOSED
```
