#!/usr/bin/env bash
set -euo pipefail

# Smoke-проверка Room Service (нужны поднятый gRPC и grpcurl).
# Использование:
#   export TOKEN='<access-jwt с user_id>'
#   export OWNER_UUID='<тот же uuid>'
#   ./scripts/smoke_rooms.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ADDR="${GRPC_ADDR:-localhost:8080}"
TOKEN="${TOKEN:-}"
OWNER_UUID="${OWNER_UUID:-}"

if ! command -v grpcurl >/dev/null 2>&1; then
  if [[ -x "${ROOT_DIR}/bin/grpcurl" ]]; then
    PATH="${ROOT_DIR}/bin:${PATH}"
  else
    echo "grpcurl не найден. Установите: task grpcurl:install (из callService)"
    exit 1
  fi
fi

echo "==> Health check on ${ADDR}"
grpcurl -plaintext "${ADDR}" grpc.health.v1.Health/Check

if [[ -z "${TOKEN}" || -z "${OWNER_UUID}" ]]; then
  echo "TOKEN/OWNER_UUID не заданы — пропускаем CreateRoom."
  echo "Сгенерируйте JWT с claim user_id=${OWNER_UUID} и секретом JWT_SECRET."
  exit 0
fi

echo "==> CreateRoom"
grpcurl -plaintext \
  -H "authorization: Bearer ${TOKEN}" \
  -d "{
    \"name\": \"smoke-room\",
    \"owner_uuid\": \"${OWNER_UUID}\",
    \"room_settings\": {
      \"max_participants\": 8,
      \"quality\": \"HD\",
      \"auto_close\": false
    }
  }" \
  "${ADDR}" rooms.v1.RoomService/CreateRoom

echo "OK smoke finished (root=${ROOT_DIR})"
