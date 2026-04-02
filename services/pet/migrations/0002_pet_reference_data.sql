-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS species (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name_ru TEXT NOT NULL,
    name_en TEXT NOT NULL,
    icon_key TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_species_sort_order ON species(sort_order, id);
CREATE INDEX IF NOT EXISTS idx_species_active_sort_order ON species(is_active, sort_order, id);

CREATE TABLE IF NOT EXISTS breeds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE RESTRICT,
    name_ru TEXT NOT NULL,
    name_en TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_breeds_species_sort_order ON breeds(species_id, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_breeds_species_active_sort_order ON breeds(species_id, is_active, sort_order, id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_breeds_species_name_en_ci ON breeds(species_id, lower(name_en));

CREATE TABLE IF NOT EXISTS patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NULL REFERENCES species(id) ON DELETE RESTRICT,
    name_ru TEXT NOT NULL,
    name_en TEXT NOT NULL,
    icon_key TEXT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patterns_species_sort_order ON patterns(species_id, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_patterns_species_active_sort_order ON patterns(species_id, is_active, sort_order, id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_patterns_species_name_en_ci
    ON patterns (COALESCE(species_id, '00000000-0000-0000-0000-000000000000'::uuid), lower(name_en));

CREATE TABLE IF NOT EXISTS color_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_ru TEXT NOT NULL,
    name_en TEXT NOT NULL,
    hex TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_color_presets_sort_order ON color_presets(sort_order, id);
CREATE INDEX IF NOT EXISTS idx_color_presets_active_sort_order ON color_presets(is_active, sort_order, id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_color_presets_name_en_ci ON color_presets(lower(name_en));

ALTER TABLE pets
    DROP CONSTRAINT IF EXISTS pets_breed_invariant,
    DROP CONSTRAINT IF EXISTS pets_coat_pattern_invariant;

ALTER TABLE pets
    DROP COLUMN IF EXISTS breed_source,
    DROP COLUMN IF EXISTS system_breed_id,
    DROP COLUMN IF EXISTS colors,
    DROP COLUMN IF EXISTS coat_pattern_source,
    DROP COLUMN IF EXISTS system_coat_pattern_id,
    DROP COLUMN IF EXISTS custom_coat_pattern_name;

ALTER TABLE pets
    ADD COLUMN IF NOT EXISTS breed_id UUID NULL REFERENCES breeds(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS pattern_id UUID NULL REFERENCES patterns(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS custom_pattern_name TEXT NULL;

ALTER TABLE pets
    ADD CONSTRAINT pets_breed_choice CHECK (
        NOT (breed_id IS NOT NULL AND custom_breed_name IS NOT NULL)
    ),
    ADD CONSTRAINT pets_pattern_choice CHECK (
        NOT (pattern_id IS NOT NULL AND custom_pattern_name IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS idx_pets_breed_id ON pets(breed_id);
CREATE INDEX IF NOT EXISTS idx_pets_pattern_id ON pets(pattern_id);

CREATE TABLE IF NOT EXISTS pet_colors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    preset_id UUID NULL REFERENCES color_presets(id) ON DELETE RESTRICT,
    custom_name TEXT NULL,
    custom_hex TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pet_colors_value CHECK (
        (preset_id IS NOT NULL AND custom_name IS NULL AND custom_hex IS NULL)
        OR
        (preset_id IS NULL AND custom_name IS NOT NULL AND custom_hex IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_pet_colors_pet_sort_order ON pet_colors(pet_id, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_pet_colors_preset_id ON pet_colors(preset_id);

-- +goose Down
DROP INDEX IF EXISTS idx_pet_colors_preset_id;
DROP INDEX IF EXISTS idx_pet_colors_pet_sort_order;
DROP TABLE IF EXISTS pet_colors;

DROP INDEX IF EXISTS idx_pets_pattern_id;
DROP INDEX IF EXISTS idx_pets_breed_id;

ALTER TABLE pets
    DROP CONSTRAINT IF EXISTS pets_pattern_choice,
    DROP CONSTRAINT IF EXISTS pets_breed_choice;

ALTER TABLE pets
    DROP COLUMN IF EXISTS custom_pattern_name,
    DROP COLUMN IF EXISTS pattern_id,
    DROP COLUMN IF EXISTS breed_id;

ALTER TABLE pets
    ADD COLUMN IF NOT EXISTS breed_source TEXT NOT NULL DEFAULT 'UNKNOWN',
    ADD COLUMN IF NOT EXISTS system_breed_id UUID NULL,
    ADD COLUMN IF NOT EXISTS colors JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS coat_pattern_source TEXT NOT NULL DEFAULT 'UNKNOWN',
    ADD COLUMN IF NOT EXISTS system_coat_pattern_id UUID NULL,
    ADD COLUMN IF NOT EXISTS custom_coat_pattern_name TEXT NULL;

ALTER TABLE pets
    ADD CONSTRAINT pets_breed_invariant CHECK (
        (breed_source = 'SYSTEM' AND system_breed_id IS NOT NULL AND custom_breed_name IS NULL) OR
        (breed_source = 'CUSTOM' AND custom_breed_name IS NOT NULL AND system_breed_id IS NULL) OR
        (breed_source = 'UNKNOWN' AND system_breed_id IS NULL AND custom_breed_name IS NULL)
    ),
    ADD CONSTRAINT pets_coat_pattern_invariant CHECK (
        (coat_pattern_source = 'SYSTEM' AND system_coat_pattern_id IS NOT NULL AND custom_coat_pattern_name IS NULL) OR
        (coat_pattern_source = 'CUSTOM' AND custom_coat_pattern_name IS NOT NULL AND system_coat_pattern_id IS NULL) OR
        (coat_pattern_source = 'UNKNOWN' AND system_coat_pattern_id IS NULL AND custom_coat_pattern_name IS NULL)
    );

DROP INDEX IF EXISTS uq_color_presets_name_en_ci;
DROP INDEX IF EXISTS idx_color_presets_active_sort_order;
DROP INDEX IF EXISTS idx_color_presets_sort_order;
DROP TABLE IF EXISTS color_presets;

DROP INDEX IF EXISTS uq_patterns_species_name_en_ci;
DROP INDEX IF EXISTS idx_patterns_species_active_sort_order;
DROP INDEX IF EXISTS idx_patterns_species_sort_order;
DROP TABLE IF EXISTS patterns;

DROP INDEX IF EXISTS uq_breeds_species_name_en_ci;
DROP INDEX IF EXISTS idx_breeds_species_active_sort_order;
DROP INDEX IF EXISTS idx_breeds_species_sort_order;
DROP TABLE IF EXISTS breeds;

DROP INDEX IF EXISTS idx_species_active_sort_order;
DROP INDEX IF EXISTS idx_species_sort_order;
DROP TABLE IF EXISTS species;
