include .env

DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
SEEDS_DIR=migrations/seeds

.PHONY: run build test migrate-up migrate-down migrate-status migrate-new psql-file seed-all seed-species seed-breeds seed-colors seed-patterns

run:
	go run ./cmd/main.go

build:
	go build -o bin/catalog ./cmd/main.go

test:
	go test ./...

migrate-up:
	goose -dir migrations postgres $(DB_URL) up

migrate-down:
	goose -dir migrations postgres $(DB_URL) down

migrate-status:
	goose -dir migrations postgres $(DB_URL) status

migrate-new:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-new name=create_catalog_tables"; exit 1; fi
	goose -dir migrations create $(name) sql

psql-file:
	@if [ -z "$(file)" ]; then echo "Usage: make psql-file file=path/to/file.sql"; exit 1; fi
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DB_URL)" -f "$(file)"; \
	else \
		docker exec -i catalog_postgres psql -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)" -f - < "$(file)"; \
	fi

seed-species:
	$(MAKE) psql-file file=$(SEEDS_DIR)/001_seed_species.sql

seed-breeds:
	$(MAKE) psql-file file=$(SEEDS_DIR)/002_seed_breeds.sql

seed-colors:
	$(MAKE) psql-file file=$(SEEDS_DIR)/003_seed_colors.sql

seed-patterns:
	$(MAKE) psql-file file=$(SEEDS_DIR)/004_seed_patterns.sql

seed-all: seed-species seed-breeds seed-colors seed-patterns
