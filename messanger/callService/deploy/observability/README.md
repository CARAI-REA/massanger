# OpenTelemetry + Grafana (local)

## Quick start (Docker Compose)

Requires `rooms-net` (core/rooms/signaling/sfu already up).

```bash
cd deploy/compose/observability
docker compose up -d
```

| URL | Creds |
|-----|--------|
| http://localhost:3000 | `admin` / `admin` |
| http://localhost:9095 | Prometheus UI (не `:9090` — это raw metrics rooms) |

**Grafana:** меню → Dashboards → папка `callService` → **callService overview**  
(или сразу http://localhost:3000/d/callservice-overview)

**Prometheus:** Graph → запрос `up` или Status → Targets (должны быть 3× UP).

Datasource Prometheus уже провижинен (`Connections` → `Data sources`), добавлять вручную не нужно.

Prometheus scrapes:

- `rooms:9090`
- `signaling:9091`
- `sfu:9093`

## OTel (optional)

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_SERVICE_NAME=rooms|signaling|sfu
OTEL_TRACES_EXPORTER=otlp
```

Prometheus `/metrics` remains the primary RED metrics source; traces complement latency debugging.
