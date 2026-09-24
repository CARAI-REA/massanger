# Kubernetes manifests (callService)

Apply order:

```bash
kubectl apply -f deploy/k8s/00-namespace-secrets.yaml
# Optional single-node Postgres/Redis for staging only:
kubectl apply -f deploy/k8s/05-infra-dev.yaml
kubectl apply -f deploy/k8s/01-rooms.yaml
kubectl apply -f deploy/k8s/02-signaling.yaml
kubectl apply -f deploy/k8s/03-sfu.yaml
kubectl apply -f deploy/k8s/04-coturn.yaml
```

Replace all `CHANGE_ME` secrets before apply. Copy [`deploy/env/.env.production.example`](../env/.env.production.example) and fill values (see [`PROD_CHECKLIST.md`](../PROD_CHECKLIST.md)).

TLS for WSS is on external nginx — see [deploy/nginx/callService.conf.example](../nginx/callService.conf.example).

## Ops notes

- **Metrics**: pods bind Prometheus to `127.0.0.1` when `APP_ENV=production` (and manifests set `METRICS_HOST=127.0.0.1`). Scrape via node-local agent / sidecar; do not expose `/metrics` on the public edge (nginx already denies).
- **SFU drain**: `preStop` POSTs `http://127.0.0.1:8082/internal/drain` then waits; new WS connects are rejected with `INSTANCE_DRAINING`.
- **Kafka**: use an external cluster with **replication factor ≥ 3** in production. Local compose keeps RF=1 for a single broker. `05-infra-dev.yaml` does **not** include Kafka.
- **Postgres / Redis**: production should use managed services. `05-infra-dev.yaml` is a single-replica stub for staging.
- **TURN**: SFU mints short-lived creds via Rooms `GetTURNCredentials` into `welcome.ice_servers`. Coturn must use `use-auth-secret` with `TURN_SHARED_SECRET`.

SFU uses `hostNetwork` so ICE UDP binds on the node. Set public TURN external IP on coturn.
