-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE species (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    icon_key TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_species_sort_order ON species(sort_order, id);
CREATE INDEX idx_species_active_sort_order ON species(is_active, sort_order, id);

CREATE TABLE breeds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NOT NULL REFERENCES species(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_breeds_species_sort_order ON breeds(species_id, sort_order, id);
CREATE INDEX idx_breeds_species_active_sort_order ON breeds(species_id, is_active, sort_order, id);
CREATE UNIQUE INDEX uq_breeds_species_name_ci ON breeds(species_id, lower(name));

CREATE TABLE patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    species_id UUID NULL REFERENCES species(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    icon_key TEXT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_patterns_species_sort_order ON patterns(species_id, sort_order, id);
CREATE INDEX idx_patterns_species_active_sort_order ON patterns(species_id, is_active, sort_order, id);
CREATE UNIQUE INDEX uq_patterns_species_name_ci
    ON patterns (COALESCE(species_id, '00000000-0000-0000-0000-000000000000'::uuid), lower(name));

CREATE TABLE color_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    hex TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_color_presets_sort_order ON color_presets(sort_order, id);
CREATE INDEX idx_color_presets_active_sort_order ON color_presets(is_active, sort_order, id);
CREATE UNIQUE INDEX uq_color_presets_name_ci ON color_presets(lower(name));

CREATE TABLE pets (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL,
    name TEXT NOT NULL,
    species_id UUID NULL REFERENCES species(id) ON DELETE RESTRICT,
    custom_species_name TEXT NULL,
    sex TEXT NOT NULL CHECK (sex IN ('FEMALE','MALE','UNKNOWN')),
    birth_date DATE NULL,

    breed_id UUID NULL REFERENCES breeds(id) ON DELETE RESTRICT,
    custom_breed_name TEXT NULL,
    pattern_id UUID NULL REFERENCES patterns(id) ON DELETE RESTRICT,
    custom_pattern_name TEXT NULL,

    is_neutered TEXT NOT NULL CHECK (is_neutered IN ('YES','NO','UNKNOWN')),
    is_outdoor BOOLEAN NOT NULL DEFAULT FALSE,

    profile_photo_file_id UUID NULL,
    microchip_id TEXT NULL,
    microchip_installed_at DATE NULL,

    status TEXT NOT NULL CHECK (status IN ('ACTIVE','MISSING','ARCHIVED')),
    missing_since TIMESTAMPTZ NULL,
    archived_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    row_version INT NOT NULL DEFAULT 1,

    CONSTRAINT pets_breed_choice CHECK (
        NOT (breed_id IS NOT NULL AND custom_breed_name IS NOT NULL)
    ),
    CONSTRAINT pets_species_choice CHECK (
        (species_id IS NOT NULL AND custom_species_name IS NULL)
        OR
        (species_id IS NULL AND custom_species_name IS NOT NULL AND length(trim(custom_species_name)) > 0)
    ),
    CONSTRAINT pets_custom_species_breed_choice CHECK (
        custom_species_name IS NULL OR breed_id IS NULL
    ),
    CONSTRAINT pets_pattern_choice CHECK (
        NOT (pattern_id IS NOT NULL AND custom_pattern_name IS NOT NULL)
    ),
    CONSTRAINT pets_archived_status_invariant CHECK (
        (status = 'ARCHIVED' AND archived_at IS NOT NULL) OR
        (status <> 'ARCHIVED' AND archived_at IS NULL)
    ),
    CONSTRAINT pets_missing_status_invariant CHECK (
        (status = 'MISSING' AND missing_since IS NOT NULL) OR
        (status <> 'MISSING' AND missing_since IS NULL)
    )
);

CREATE INDEX idx_pets_owner_user_id ON pets(owner_user_id);
CREATE INDEX idx_pets_status ON pets(status);
CREATE INDEX idx_pets_species_id ON pets(species_id);
CREATE INDEX idx_pets_breed_id ON pets(breed_id);
CREATE INDEX idx_pets_pattern_id ON pets(pattern_id);

CREATE TABLE pet_colors (
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

CREATE INDEX idx_pet_colors_pet_sort_order ON pet_colors(pet_id, sort_order, id);
CREATE INDEX idx_pet_colors_preset_id ON pet_colors(preset_id);

-- +goose Down
DROP TABLE IF EXISTS pet_colors;
DROP TABLE IF EXISTS pets;
DROP TABLE IF EXISTS color_presets;
DROP TABLE IF EXISTS patterns;
DROP TABLE IF EXISTS breeds;
DROP TABLE IF EXISTS species;
