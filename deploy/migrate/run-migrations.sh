#!/bin/sh
set -eu

usage() {
  echo "Usage: run-migrations <service|all> <up|down|status>"
  echo "Services: auth file acl profile pet health push chat"
}

internal_host() {
  case "$1" in
    auth) echo "postgres" ;;
    file) echo "file_postgres" ;;
    acl) echo "acl_postgres" ;;
    profile) echo "profile_postgres" ;;
    pet) echo "pet_postgres" ;;
    health) echo "health_postgres" ;;
    push) echo "push_postgres" ;;
    chat) echo "chat_postgres" ;;
    *)
      echo "unknown service: $1" >&2
      exit 1
      ;;
  esac
}

run_one() {
  service="$1"
  cmd="$2"
  env_file="/work/services/$service/.goose.env"
  migrations_dir="/work/services/$service/migrations"

  if [ ! -f "$env_file" ]; then
    echo "missing env file: $env_file" >&2
    exit 1
  fi
  if [ ! -d "$migrations_dir" ]; then
    echo "missing migrations dir: $migrations_dir" >&2
    exit 1
  fi

  unset POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_HOST POSTGRES_PORT
  set -a
  . "$env_file"
  set +a

  host="$(internal_host "$service")"
  port="5432"
  db_url="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${host}:${port}/${POSTGRES_DB}?sslmode=disable"

  echo "==> $service: goose $cmd"
  goose -dir "$migrations_dir" postgres "$db_url" "$cmd"
}

target="${1:-}"
cmd="${2:-up}"

if [ -z "$target" ]; then
  usage
  exit 1
fi

case "$cmd" in
  up|down|status) ;;
  *)
    usage
    exit 1
    ;;
esac

if [ "$target" = "all" ]; then
  for service in auth file acl profile pet health push chat; do
    run_one "$service" "$cmd"
  done
else
  run_one "$target" "$cmd"
fi
