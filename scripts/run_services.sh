#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

services=(
  "auth"
  "acl"
  "pet"
  "catalog"
  "notification"
  "email"
  "file"
  "profile"
)

pids=()
failed=()

start_service() {
  local name="$1"
  local dir="$ROOT_DIR/services/$name"
  local env_file="$dir/.env"

  if [[ ! -f "$env_file" ]]; then
    echo "Skipping $name: missing .env"
    failed+=("$name")
    return
  fi

  echo "Starting $name..."
  (
    cd "$dir"
    set -a
    source "$env_file"
    set +a
    go run ./cmd/main.go
  ) &
  local pid="$!"
  pids+=("$pid")

  sleep 0.8
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "Failed to start $name (process exited early)"
    failed+=("$name")
  fi
}

for svc in "${services[@]}"; do
  start_service "$svc"
  sleep 0.2
 done

if [[ ${#failed[@]} -gt 0 ]]; then
  echo "Failed services: ${failed[*]}"
  echo "Stopping started processes..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  exit 1
fi

echo "All services started. PIDs: ${pids[*]}"

echo "Press Ctrl+C to stop."
trap 'echo "Stopping..."; kill "${pids[@]}"' INT TERM
wait
