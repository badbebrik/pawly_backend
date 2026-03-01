-- +goose Up
CREATE TABLE pets (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL,
    name TEXT NOT NULL,
    species_id UUID NOT NULL,
    sex TEXT NOT NULL CHECK (sex IN ('FEMALE','MALE','UNKNOWN')),
    birth_date DATE NULL,

    breed_source TEXT NOT NULL CHECK (breed_source IN ('SYSTEM','CUSTOM','UNKNOWN')),
    system_breed_id UUID NULL,
    custom_breed_name TEXT NULL,

    colors JSONB NOT NULL DEFAULT '[]'::jsonb,

    coat_pattern_source TEXT NOT NULL CHECK (coat_pattern_source IN ('SYSTEM','CUSTOM','UNKNOWN')),
    system_coat_pattern_id UUID NULL,
    custom_coat_pattern_name TEXT NULL,

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

    CONSTRAINT pets_breed_invariant CHECK (
        (breed_source = 'SYSTEM' AND system_breed_id IS NOT NULL AND custom_breed_name IS NULL) OR
        (breed_source = 'CUSTOM' AND custom_breed_name IS NOT NULL AND system_breed_id IS NULL) OR
        (breed_source = 'UNKNOWN' AND system_breed_id IS NULL AND custom_breed_name IS NULL)
    ),
    CONSTRAINT pets_coat_pattern_invariant CHECK (
        (coat_pattern_source = 'SYSTEM' AND system_coat_pattern_id IS NOT NULL AND custom_coat_pattern_name IS NULL) OR
        (coat_pattern_source = 'CUSTOM' AND custom_coat_pattern_name IS NOT NULL AND system_coat_pattern_id IS NULL) OR
        (coat_pattern_source = 'UNKNOWN' AND system_coat_pattern_id IS NULL AND custom_coat_pattern_name IS NULL)
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

-- +goose Down
DROP TABLE IF EXISTS pets;
