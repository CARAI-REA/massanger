# SFU Media Server — TZ checklist

## MVP + production

- [x] WebSocket `/v1/ws` + join JWT
- [x] AssertCanJoin (service JWT)
- [x] Pion offer/answer/ICE + renegotiation
- [x] Track fan-out + late joiner
- [x] TURN / NAT1To1 / client ICE override
- [x] Redis affinity + Kafka force-close
- [x] APP_ENV fail-fast / WS query token gate
- [x] Recording egress (feature flag + RTP dump)
- [x] Simulcast RID filter (feature flag)
- [x] OTel optional + Prometheus
- [x] k8s DaemonSet (hostNetwork) + nginx example
- [x] PROTOCOL.md

## Notes

- Production TURN: coturn `use-auth-secret` + Rooms `GetTURNCredentials`
- Local compose may still use static TURN user for convenience
