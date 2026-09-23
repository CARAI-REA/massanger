#!/usr/bin/env bash
# Prepare compose .env files for CI (templates are committed; .env is gitignored).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ ! -f deploy/env/.env ]]; then
  cp deploy/env/.env.template deploy/env/.env
fi

# Force development so prodguard does not block containers / env:generate.
sed -i.bak -E \
  -e 's/^(ROOMS_APP_ENV|SIGNALING_APP_ENV|SFU_APP_ENV)=.*/\1=development/' \
  deploy/env/.env || true
rm -f deploy/env/.env.bak

if command -v task >/dev/null 2>&1; then
  task env:generate
  exit 0
fi

# Fallback without task: simple envsubst per service if available
if command -v envsubst >/dev/null 2>&1; then
  set -a
  # shellcheck disable=SC1091
  source deploy/env/.env
  set +a
  for svc in core rooms signaling sfu; do
    mkdir -p "deploy/compose/${svc}"
    envsubst < "deploy/env/${svc}.env.template" > "deploy/compose/${svc}/.env"
  done
  exit 0
fi

echo "Neither task nor envsubst found; copying committed compose .env if present"
for svc in core rooms signaling sfu; do
  if [[ ! -f "deploy/compose/${svc}/.env" && -f "deploy/env/${svc}.env.template" ]]; then
    echo "WARN: ${svc} .env missing — install gettext envsubst or task"
  fi
done
