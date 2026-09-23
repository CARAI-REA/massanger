#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HTTP_ADDR="${SFU_ADDR:-localhost:8082}"

echo "==> Health ${HTTP_ADDR}/healthz"
curl -fsS "http://${HTTP_ADDR}/healthz" >/dev/null
echo "OK healthz"

echo "==> Ready ${HTTP_ADDR}/readyz"
curl -fsS "http://${HTTP_ADDR}/readyz" >/dev/null || echo "WARN readyz (redis?)"
echo "==> Metrics"
curl -fsS "http://localhost:9093/metrics" | head -n 5 || true

echo "For full E2E: cd sfu && go run ./cmd/e2e"
echo "OK smoke finished (root=${ROOT_DIR})"
