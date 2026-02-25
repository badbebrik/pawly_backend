#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

services=(
  "auth"
  "acl"
  "catalog"
  "notification"
  "email"
  "file"
  "profile"
)

pids=()

start_service() {
  local name="$1"
  local dir="$ROOT_DIR/services/$name"
  echo "Starting $name..."
  (
    cd "$dir"
    set -a
    source .env
    set +a
    go run ./cmd/main.go
  ) &
  pids+=("$!")
}

for svc in "${services[@]}"; do
  start_service "$svc"
  sleep 0.5
 done

echo "All services started. PIDs: ${pids[*]}"

echo "Press Ctrl+C to stop."
trap 'echo "Stopping..."; kill "${pids[@]}"' INT TERM
wait
