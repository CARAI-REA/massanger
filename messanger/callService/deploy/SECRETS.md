# Secrets inventory (callService)

Do **not** commit real values. Use Vault / Kubernetes Secrets / sealed-secrets in production.
`APP_ENV=production` refuses known weak/dev defaults at process start.

## Required secrets

| Secret | Used by | Notes |
|--------|---------|-------|
| `JWT_SECRET` / `ROOMS_JWT_SECRET` | rooms | User access JWT (API gateway) |
| `JOIN_TOKEN_SECRET_KEY` | rooms, signaling, sfu | Must be identical across all three |
| `SERVICE_JWT_SECRET` | rooms, signaling, sfu | Internal service-to-service JWT |
| `POSTGRES_PASSWORD` | rooms | Strong password; SSL required in prod |
| `REDIS_PASSWORD` | rooms, signaling, sfu | `requirepass` on Redis |
| `WEBRTC_TURN_CREDENTIAL` / `TURN_SHARED_SECRET` | rooms (HMAC), sfu, coturn | coturn `use-auth-secret` static-auth-secret |
| Kafka credentials (if enabled) | all | Cluster-specific |

## Rotation

1. Issue new secret in Vault.
2. Roll pods with dual-read window if supported; otherwise short maintenance.
3. Invalidate old join tokens by rotating `JOIN_TOKEN_SECRET_KEY` (forces re-join).
4. Rotate TURN shared secret together with Rooms `TURN_SHARED_SECRET` and coturn.

## Local vs production

| | Local (`APP_ENV=development`) | Production |
|--|-------------------------------|------------|
| Secrets | Weak placeholders allowed | Fail-fast on weak/localhost |
| Postgres SSL | `disable` | `require` / `verify-full` |
| Redis AUTH | optional | required |
| WS query `?token=` | allowed | disabled (`WS_ALLOW_QUERY_TOKEN=false`) |
| Origins | empty = allow all | must be set |

See `deploy/env/.env.template` for local variable names and `deploy/env/.env.production.example` + `deploy/PROD_CHECKLIST.md` for production.
