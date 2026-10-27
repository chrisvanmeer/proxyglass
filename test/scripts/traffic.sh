#!/usr/bin/env bash
set -euo pipefail
count="${1:-100}"
for port in 18080 28080 38080; do
  for _ in $(seq 1 "$count"); do
    curl --silent --output /dev/null --max-time 2 "http://localhost:${port}/" || true
  done
done
echo "Sent up to $((count * 3)) requests through the test HAProxy frontends."
