#!/usr/bin/env bash
# Rooms gRPC load (requires ghz: https://github.com/bojand/ghz)
set -euo pipefail
ADDR="${ROOMS_ADDR:-localhost:8080}"
TOKEN="${TOKEN:?set TOKEN=access JWT}"
OWNER="${OWNER_UUID:?set OWNER_UUID}"
ghz --insecure \
  --proto ../../shared/proto/rooms/v1/rooms.proto \
  --call rooms.v1.RoomService/GetRoom \
  -d "{\"room_uuid\":\"${ROOM_UUID:-00000000-0000-0000-0000-000000000001}\"}" \
  -m "{\"authorization\":\"Bearer ${TOKEN}\"}" \
  -n 1000 -c 50 \
  "$ADDR"
