# SFU Media Server — TZ checklist

## MVP + production

- [x] WebSocket `/v1/ws` + join JWT
- [x] AssertCanJoin (service JWT) + recording_enabled flag
- [x] Pion offer/answer/ICE + renegotiation
- [x] Track fan-out + late joiner
- [x] Ephemeral TURN in welcome via Rooms GetTURNCredentials
- [x] NAT1To1 / client ICE STUN merge
- [x] Redis affinity + Kafka force-close
- [x] APP_ENV fail-fast / WS_ALLOW_QUERY_TOKEN=false required in prod
- [x] TOKEN_EXPIRED WS close 4001
- [x] Drain (`/internal/drain` + SIGTERM) / INSTANCE_DRAINING
- [x] Recording: per-room flag + global env → RTP dump (not HLS/S3)
- [x] Simulcast RID filter (feature flag) — **SVC / BWE: future**
- [x] OTel tracer + otelgrpc + minimal session spans
- [x] Metrics localhost bind in production
- [x] k8s DaemonSet (hostNetwork) + nginx example
- [x] PROTOCOL.md

## Out of scope / future

- HLS/WebM/S3 recording pipeline
- Simulcast SVC / bandwidth adaptation
- Multi-broker Kafka provisioning (use external RF≥3 in prod)
