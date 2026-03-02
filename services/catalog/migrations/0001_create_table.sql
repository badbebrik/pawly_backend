-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS catalog_version (
   id      SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
   version INT NOT NULL
);

INSERT INTO catalog_version (id, version)
VALUES (1, 1)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS species (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ru    TEXT NOT NULL,
    name_en    TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    version    INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS breeds (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE RESTRICT,
    name_ru    TEXT NOT NULL,
    name_en    TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    version    INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_breeds_species_id ON breeds(species_id);

CREATE TABLE IF NOT EXISTS colors (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ru    TEXT NOT NULL,
    name_en    TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    version    INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS patterns (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ru    TEXT NOT NULL,
    name_en    TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    version    INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS breeds;
DROP TABLE IF EXISTS colors;
DROP TABLE IF EXISTS patterns;
DROP TABLE IF EXISTS species;
DROP TABLE IF EXISTS catalog_version;
