# Room Service — чеклист ТЗ (`room_task.md`)

## R / P / A / E

Все пункты MVP закрыты (см. историю). Дополнительно для prod:

| ID | Статус |
|----|--------|
| AssertCanJoin (service JWT) | ✅ |
| RefreshJoinToken | ✅ |
| GetTURNCredentials (HMAC) | ✅ |
| recording_enabled в proto/model/SQL | ✅ |
| APP_ENV fail-fast | ✅ |
| Postgres SSL / Redis AUTH | ✅ |
| gRPC rate limit | ✅ |
| OpenTelemetry (OTLP optional) | ✅ |
| Load scripts (ghz) | ✅ |
| K8s manifests | ✅ |

## Как считать «готово»

- [x] Prod secrets validation
- [x] Service + user auth paths
- [x] Join token refresh
- [x] Metrics private (nginx deny)
- [x] Docs: PROTOCOL.md, SECRETS.md
