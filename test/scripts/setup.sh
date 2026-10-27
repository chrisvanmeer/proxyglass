#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
docker compose -f docker/docker-compose.test.yaml up -d --build --force-recreate

ready=false
printf 'Waiting for three HAProxy stats endpoints'
for _ in $(seq 1 45); do
  if curl -fsS -u admin:changeme 'http://localhost:18404/stats;csv' >/dev/null 2>&1 && \
     curl -fsS -u admin:changeme 'http://localhost:28404/stats;csv' >/dev/null 2>&1 && \
     curl -fsS -u admin:changeme 'http://localhost:38404/stats;csv' >/dev/null 2>&1; then
    ready=true
    break
  fi
  printf '.'
  sleep 1
done
if [[ "$ready" != true ]]; then
  echo '\nHAProxy stats did not become ready. Inspect logs with docker compose -f docker/docker-compose.test.yaml logs haproxy1 haproxy2 haproxy3.' >&2
  exit 1
fi
app_ready=false
for _ in $(seq 1 30); do
  if curl -fsS http://localhost:18000/healthz >/dev/null 2>&1; then
    app_ready=true
    break
  fi
  sleep 1
done
if [[ "$app_ready" != true ]]; then
  echo 'Proxyglass did not become ready. Inspect logs with docker compose -f docker/docker-compose.test.yaml logs proxyglass.' >&2
  exit 1
fi
printf '\nGenerating sample requests through each HAProxy frontend...\n'
"$ROOT/test/scripts/traffic.sh" 150
token="$(docker compose -f docker/docker-compose.test.yaml exec -T proxyglass /usr/bin/proxyglass --generate-token --quiet --config /etc/proxyglass/config.yml)"

printf '\nDashboard: http://localhost:18000\n'
printf 'Test login token (paste into the login page): %s\n' "$token"
printf 'Stats pages (basic auth admin / changeme):\n'
printf '  http://localhost:18404/stats\n  http://localhost:28404/stats\n  http://localhost:38404/stats\n'
