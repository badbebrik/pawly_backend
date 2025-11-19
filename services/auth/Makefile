include .goose.env

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

migrate-up:
	goose -dir migrations postgres $(DB_URL) up

migrate-down:
	goose -dir migrations postgres $(DB_URL) down

migrate-status:
	goose -dir migrations postgres $(DB_URL) status

migrate-new:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-new name=create_users"; exit 1; fi
	goose -dir migrations create $(name) sql