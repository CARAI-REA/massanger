# Signaling Service — TZ checklist

## MVP + production

- [x] WebSocket `/v1/ws`
- [x] Join JWT (Bearer; query token gated; prod requires `WS_ALLOW_QUERY_TOKEN=false`)
- [x] TOKEN_EXPIRED close code 4001
- [x] Room gRPC `AssertCanJoin` + service JWT
- [x] welcome / peer_joined / peer_left / bye / SESSION_REPLACED
- [x] offer / answer / ice_candidate relay (mesh fallback)
- [x] Redis presence + Pub/Sub
- [x] Kafka force-close
- [x] Rate limit / Origin allowlist / APP_ENV fail-fast
- [x] Prometheus (localhost in prod) + otelgrpc client
- [x] SFU welcome hint
- [x] Load script (k6; not in CI) + CI unit job

## Out of scope

- Media path (SFU)
- Recording (SFU)
