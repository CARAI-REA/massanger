#!/usr/bin/env bash
set -euo pipefail

# Smoke for Signaling: health + optional WS ping with JOIN_TOKEN.
# Usage:
#   JOIN_TOKEN=... ROOM_UUID=... ./scripts/smoke_signaling.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HTTP_ADDR="${SIGNALING_ADDR:-localhost:8081}"
WS_PATH="${WS_PATH:-/v1/ws}"
JOIN_TOKEN="${JOIN_TOKEN:-}"
ROOM_UUID="${ROOM_UUID:-}"

echo "==> Health ${HTTP_ADDR}/healthz"
curl -fsS "http://${HTTP_ADDR}/healthz" >/dev/null
echo "OK healthz"

echo "==> Metrics ${HTTP_ADDR%:*}:9091/metrics (best-effort)"
METRICS_HOST="${HTTP_ADDR%%:*}"
curl -fsS "http://${METRICS_HOST}:9091/metrics" | head -n 5 || true

if [[ -z "${JOIN_TOKEN}" || -z "${ROOM_UUID}" ]]; then
  echo "JOIN_TOKEN/ROOM_UUID not set — skipping WS check."
  echo "Create room via rooms, take join_token, then re-run with env set."
  exit 0
fi

if ! command -v websocat >/dev/null 2>&1; then
  echo "websocat not found — install to exercise WS (optional)."
  exit 0
fi

echo "==> WS ping"
URL="ws://${HTTP_ADDR}${WS_PATH}?room_uuid=${ROOM_UUID}&token=${JOIN_TOKEN}"
# send ping envelope and expect pong
printf '{"type":"ping","payload":{}}\n' | websocat -t -1 "$URL" | head -n 5
echo "OK smoke finished (root=${ROOT_DIR})"
