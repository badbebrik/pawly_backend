-include .goose.env

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
GOCACHE?=/tmp/push-go-build-cache

.PHONY: run test build migrate migrate-up migrate-down migrate-status migrate-new up down status check-db-env

run:
	GOCACHE=$(GOCACHE) go run ./cmd/main.go

test:
	GOCACHE=$(GOCACHE) go test ./...

build:
	GOCACHE=$(GOCACHE) go build -o bin/push ./cmd/main.go

migrate-up: check-db-env
	goose -dir migrations postgres $(DB_URL) up

migrate-down: check-db-env
	goose -dir migrations postgres $(DB_URL) down

migrate-status: check-db-env
	goose -dir migrations postgres $(DB_URL) status

migrate-new:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-new name=create_push_tables"; exit 1; fi
	goose -dir migrations create $(name) sql

migrate:
	@true

up: migrate-up

down: migrate-down

status: migrate-status
