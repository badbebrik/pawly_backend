-- +goose Up
CREATE TABLE vet_visits (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'COMPLETED', 'CANCELLED')),
    visit_type TEXT NOT NULL CHECK (visit_type IN ('CHECKUP', 'SYMPTOM', 'FOLLOW_UP', 'VACCINATION', 'PROCEDURE', 'OTHER')),
    scheduled_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    reason_text TEXT NULL,
    result_text TEXT NULL,
    clinic_name TEXT NULL,
    vet_name TEXT NULL,
    row_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL
);

CREATE INDEX idx_vet_visits_pet_active ON vet_visits (pet_id, updated_at DESC, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_vet_visits_pet_status_active ON vet_visits (pet_id, status, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_vet_visits_pet_scheduled_active ON vet_visits (pet_id, scheduled_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE TABLE entity_relations (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    left_entity_type TEXT NOT NULL CHECK (left_entity_type IN ('LOG', 'VACCINATION', 'PROCEDURE', 'MEDICAL_RECORD')),
    left_entity_id UUID NOT NULL,
    right_entity_type TEXT NOT NULL CHECK (right_entity_type IN ('VET_VISIT', 'PROCEDURE', 'VACCINATION', 'MEDICAL_RECORD')),
    right_entity_id UUID NOT NULL,
    created_by_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (left_entity_type, left_entity_id, right_entity_type, right_entity_id)
);

CREATE INDEX idx_entity_relations_left ON entity_relations (pet_id, left_entity_type, left_entity_id);
CREATE INDEX idx_entity_relations_right ON entity_relations (pet_id, right_entity_type, right_entity_id);

CREATE TABLE vaccinations (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'DONE', 'CANCELLED')),
    vaccine_name TEXT NOT NULL,
    catalog_medication_id UUID NULL,
    scheduled_at TIMESTAMPTZ NULL,
    administered_at TIMESTAMPTZ NULL,
    next_due_at TIMESTAMPTZ NULL,
    clinic_name TEXT NULL,
    vet_name TEXT NULL,
    notes TEXT NULL,
    row_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL
);

CREATE INDEX idx_vaccinations_pet_active ON vaccinations (pet_id, updated_at DESC, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccinations_pet_status_active ON vaccinations (pet_id, status, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccinations_pet_scheduled_active ON vaccinations (pet_id, scheduled_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE TABLE procedures (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'DONE', 'CANCELLED')),
    procedure_type TEXT NOT NULL CHECK (procedure_type IN ('PARASITE_TREATMENT', 'DEWORMING', 'HYGIENE', 'WOUND_CARE', 'GROOMING', 'OTHER')),
    title TEXT NOT NULL,
    description TEXT NULL,
    catalog_medication_id UUID NULL,
    product_name TEXT NULL,
    scheduled_at TIMESTAMPTZ NULL,
    performed_at TIMESTAMPTZ NULL,
    next_due_at TIMESTAMPTZ NULL,
    notes TEXT NULL,
    row_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL
);

CREATE INDEX idx_procedures_pet_active ON procedures (pet_id, updated_at DESC, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_procedures_pet_status_active ON procedures (pet_id, status, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_procedures_pet_scheduled_active ON procedures (pet_id, scheduled_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE TABLE medical_records (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    record_type TEXT NOT NULL CHECK (record_type IN ('DIAGNOSIS', 'ALLERGY', 'CHRONIC_CONDITION', 'INJURY', 'SURGERY', 'CLINICAL_NOTE')),
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'RESOLVED')),
    title TEXT NOT NULL,
    description TEXT NULL,
    started_at TIMESTAMPTZ NULL,
    resolved_at TIMESTAMPTZ NULL,
    row_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL
);

CREATE INDEX idx_medical_records_pet_active ON medical_records (pet_id, updated_at DESC, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_medical_records_pet_status_active ON medical_records (pet_id, status, id DESC)
WHERE deleted_at IS NULL;
CREATE INDEX idx_medical_records_pet_started_active ON medical_records (pet_id, started_at DESC, id DESC)
WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS medical_records;
DROP TABLE IF EXISTS procedures;
DROP TABLE IF EXISTS vaccinations;
DROP TABLE IF EXISTS entity_relations;
DROP TABLE IF EXISTS vet_visits;
