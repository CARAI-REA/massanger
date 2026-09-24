# Local access (Docker Compose)

Доступы **с хоста** (IDE, GUI-клиенты, браузер). Стек: Colima/Docker Compose, сеть `rooms-net`.

> С хоста всегда используй **`127.0.0.1`** или **`localhost`**.  
> Имена контейнеров (`postgres-rooms`, `redis`, `kafka`, …) резолвятся **только внутри** Docker-сети.

Источник дефолтов: `deploy/env/.env.template`. Это **dev**-секреты, не для prod.

---

## Быстрая таблица портов (хост)

| Сервис | Порт | URL / endpoint |
|--------|------|----------------|
| Rooms gRPC | `8080` | `127.0.0.1:8080` |
| Signaling HTTP/WS | `8081` | http://127.0.0.1:8081 · `ws://127.0.0.1:8081/v1/ws` |
| SFU HTTP/WS | `8082` | http://127.0.0.1:8082 · `ws://127.0.0.1:8082/v1/ws` |
| Postgres | `5432` | см. ниже |
| Redis | `6379` | см. ниже |
| Kafka (host) | `9092` | `127.0.0.1:9092` |
| Kafka UI | `8090` | http://127.0.0.1:8090 |
| Rooms metrics | `9090` | http://127.0.0.1:9090/metrics |
| Signaling metrics | `9091` | http://127.0.0.1:9091/metrics |
| SFU metrics | `9093` | http://127.0.0.1:9093/metrics |
| Prometheus | `9095` | http://127.0.0.1:9095 |
| Grafana | `3000` | http://127.0.0.1:3000 |
| coturn TURN/STUN | `3478` | `turn:127.0.0.1:3478?transport=udp` |
| SFU ICE UDP | `10000–10100` | UDP |

Kafka внутри Docker: `kafka:29092`. Controller: `29093` (не для клиентов с хоста).

---

## PostgreSQL

| | |
|---|---|
| Host (с хоста) | `127.0.0.1` |
| Host (из контейнера) | `postgres-rooms` |
| Port | `5432` |
| User | `rooms_user` |
| Password | `rooms_password` |
| Database | `rooms` |
| SSL | `disable` |

```
postgresql://rooms_user:rooms_password@127.0.0.1:5432/rooms?sslmode=disable
```

---

## Redis

| | |
|---|---|
| Host (с хоста) | `127.0.0.1` |
| Host (из контейнера) | `redis` |
| Port | `6379` |
| Password | *(пусто)* |

```
redis://127.0.0.1:6379/0
```

---

## Kafka

| | |
|---|---|
| Brokers (с хоста) | `127.0.0.1:9092` |
| Brokers (из контейнера) | `kafka:29092` |
| Topics | `room-events`, `participant-events` |
| UI | http://127.0.0.1:8090 (без логина) |

---

## Grafana / Prometheus

| | |
|---|---|
| Grafana | http://127.0.0.1:3000 |
| User / Password | `admin` / `admin` |
| Dashboard | Dashboards → `callService` → **callService overview** |
| Прямая ссылка | http://127.0.0.1:3000/d/callservice-overview |
| Prometheus | http://127.0.0.1:9095 |
| Datasource | уже провижинен (`Prometheus` → `http://prometheus:9090`) |

---

## Rooms / Signaling / SFU

| | Rooms | Signaling | SFU |
|---|--------|-----------|-----|
| API | gRPC `:8080` | HTTP/WS `:8081` | HTTP/WS `:8082` |
| Metrics | `:9090/metrics` | `:9091/metrics` | `:9093/metrics` |
| WS path | — | `/v1/ws` | `/v1/ws` |

### JWT / secrets (local)

| Переменная | Значение |
|------------|----------|
| `JWT_SECRET` (API Gateway auth) | `rooms-auth-secret` |
| `JOIN_TOKEN_SECRET_KEY` | `rooms-join-token-secret` |
| `SERVICE_JWT_SECRET` | `rooms-service-jwt-secret` |
| Join token TTL | `5m` |

---

## TURN / WebRTC (coturn)

| | |
|---|---|
| URL | `turn:127.0.0.1:3478?transport=udp` |
| Username | `sfu` |
| Credential / shared secret | `sfu-turn-secret` |
| STUN (Google) | `stun:stun.l.google.com:19302` |
| Relay UDP | `49160–49200` |
| SFU media UDP | `10000–10100` |

---

## Имена контейнеров (только Docker-сеть)

| Контейнер | Роль |
|-----------|------|
| `postgres-rooms` | Postgres |
| `redis` | Redis |
| `kafka` / `kafka-ui` | Kafka |
| `rooms` / `signaling` / `sfu` | сервисы |
| `sfu-coturn` | TURN |
| `callservice-prometheus` / `callservice-grafana` | observability |
