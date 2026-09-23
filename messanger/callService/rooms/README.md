# Room Service

Микросервис управления комнатами и участниками видеозвонков (gRPC).

## Быстрый старт

```bash
# 1. Сгенерировать env (из корня callService)
task env:generate

# 2. Поднять инфраструктуру и сервис
task up-rooms
# эквивалент: task up-core && task up-rooms-postgres
```

Локально без Docker-образа сервиса:

```bash
task up-core
task up-rooms-postgres   # поднимет postgres (+ rooms image, если в compose)
cd rooms && go run ./cmd
```

Конфиг по умолчанию: `deploy/compose/rooms/.env`.

## Порты

| Сервис | Порт |
|--------|------|
| gRPC RoomService | `8080` (`GRPC_PORT`) |
| Prometheus metrics | `9090` (`METRICS_PORT`) → `/metrics` |
| Postgres | `5432` |
| Redis | `6379` |
| Kafka | `9092` |

## Auth

Все RPC (кроме health/reflection) требуют:

```text
authorization: Bearer <JWT>
```

JWT подписывается `JWT_SECRET`, claim: `"user_id": "<uuid>"`.

Join-token для Signaling выдаётся в ответах CreateRoom / AddParticipant (`JOIN_TOKEN_*`).

## Smoke (grpcurl)

```bash
# health
grpcurl -plaintext localhost:8080 grpc.health.v1.Health/Check

# пример CreateRoom (подставьте свой access JWT)
export TOKEN='...'
grpcurl -plaintext \
  -H "authorization: Bearer ${TOKEN}" \
  -d '{
    "name": "demo",
    "owner_uuid": "<same-as-user_id-in-jwt>",
    "room_settings": {"max_participants": 8, "quality": "HD", "auto_close": false}
  }' \
  localhost:8080 rooms.v1.RoomService/CreateRoom
```

Скрипт: `scripts/smoke_rooms.sh`.

## Тесты

```bash
cd rooms

# unit
go test ./... -short

# unit + integration (нужен Docker для testcontainers)
go test ./...
```

## Чеклист ТЗ

См. [TZ_CHECKLIST.md](./TZ_CHECKLIST.md).
