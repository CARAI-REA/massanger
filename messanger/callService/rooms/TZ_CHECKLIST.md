# Room Service — чеклист ТЗ (`room_task.md`)

## R / P / A / E

| ID | Статус |
|----|--------|
| AssertCanJoin (service JWT) + recording_enabled | ✅ |
| RefreshJoinToken | ✅ |
| GetTURNCredentials (HMAC; called by SFU) | ✅ |
| recording_enabled в proto/model/SQL | ✅ (SFU RTP dump; HLS/S3 — future) |
| APP_ENV fail-fast | ✅ |
| Postgres SSL / Redis AUTH | ✅ |
| gRPC rate limit | ✅ |
| OpenTelemetry (OTLP + otelgrpc) | ✅ |
| Load scripts (ghz; not in CI) | ✅ |
| K8s manifests | ✅ (templates; CHANGE_ME) |

## Как считать «готово»

- [x] Prod secrets validation
- [x] Service + user auth paths
- [x] Join token refresh
- [x] Metrics localhost in prod + nginx deny
- [x] Docs: PROTOCOL.md, SECRETS.md
