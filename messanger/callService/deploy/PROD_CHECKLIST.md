# Production readiness checklist (callService)

Secrets you fill yourself — do **not** commit real values. Use [`deploy/env/.env.production.example`](env/.env.production.example) and [`SECRETS.md`](SECRETS.md).

## You must provide

- [ ] Copy [`deploy/env/.env.production`](env/.env.production) (gitignored starter secrets) or generate new ones
- [ ] Rotate all secrets before real go-live
- [ ] Set real hosts: Postgres, Redis, Kafka, TURN URLs, `WEBRTC_NAT_1TO1_IPS`, WSS origins
- [ ] For k8s: `kubectl apply -f deploy/k8s/00-namespace-secrets.local.yaml` (gitignored) then the rest of manifests


## Already done in code / manifests

- [x] Join JWT + AssertCanJoin + RefreshJoinToken
- [x] Ephemeral TURN minted into SFU `welcome.ice_servers`
- [x] Prod fail-fast (weak secrets, query token, origins, NAT1To1)
- [x] TOKEN_EXPIRED close `4001`, SFU drain + `INSTANCE_DRAINING`
- [x] Metrics bind `127.0.0.1` in production
- [x] Recording: room `recording_enabled` → RTP dump (not HLS/S3)
- [x] otelgrpc + optional OTLP
- [x] Compose stack + e2e script (offer/answer + force-close)
- [x] k8s app manifests + optional [`05-infra-dev.yaml`](k8s/05-infra-dev.yaml) for local Postgres/Redis

## Out of scope (future)

- HLS / S3 recording pipeline
- Simulcast SVC / BWE
- Managed multi-broker Kafka provisioning

## Verify before go-live

```bash
# local compose
cd scripts/e2e_full_stack && go run .

# production config dry-run
APP_ENV=production ... # process must refuse weak secrets / query token
```
