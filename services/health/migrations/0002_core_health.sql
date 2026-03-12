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

CREATE TABLE vet_visit_log_refs (
    id UUID PRIMARY KEY,
    vet_visit_id UUID NOT NULL REFERENCES vet_visits(id) ON DELETE CASCADE,
    log_id UUID NOT NULL REFERENCES logs(id) ON DELETE CASCADE,
    added_by_user_id UUID NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (vet_visit_id, log_id)
);

CREATE INDEX idx_vet_visit_log_refs_log_id ON vet_visit_log_refs (log_id);

CREATE TABLE vaccinations (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'DONE', 'CANCELLED')),
    vaccine_name TEXT NOT NULL,
    catalog_medication_id UUID NULL,
    scheduled_at TIMESTAMPTZ NULL,
    administered_at TIMESTAMPTZ NULL,
    next_due_at TIMESTAMPTZ NULL,
    vet_visit_id UUID NULL REFERENCES vet_visits(id),
    clinic_name TEXT NULL,
    vet_name TEXT NULL,
    notes TEXT NULL,
    source_vaccination_id UUID NULL REFERENCES vaccinations(id),
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
CREATE UNIQUE INDEX uq_vaccinations_source_planned_active ON vaccinations (source_vaccination_id)
WHERE source_vaccination_id IS NOT NULL AND deleted_at IS NULL AND status = 'PLANNED';

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
    vet_visit_id UUID NULL REFERENCES vet_visits(id),
    notes TEXT NULL,
    source_procedure_id UUID NULL REFERENCES procedures(id),
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
CREATE UNIQUE INDEX uq_procedures_source_planned_active ON procedures (source_procedure_id)
WHERE source_procedure_id IS NOT NULL AND deleted_at IS NULL AND status = 'PLANNED';

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

CREATE TABLE health_attachment_refs (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('VET_VISIT', 'VACCINATION', 'PROCEDURE', 'MEDICAL_RECORD')),
    entity_id UUID NOT NULL,
    file_id UUID NOT NULL,
    file_name TEXT NULL,
    file_type TEXT NOT NULL CHECK (file_type IN ('image', 'pdf', 'other')),
    added_by_user_id UUID NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (entity_type, entity_id, file_id)
);

CREATE INDEX idx_health_attachment_refs_entity ON health_attachment_refs (entity_type, entity_id, added_at DESC);
CREATE INDEX idx_health_attachment_refs_file ON health_attachment_refs (file_id);

-- +goose Down
DROP TABLE IF EXISTS health_attachment_refs;
DROP TABLE IF EXISTS medical_records;
DROP TABLE IF EXISTS procedures;
DROP TABLE IF EXISTS vaccinations;
DROP TABLE IF EXISTS vet_visit_log_refs;
DROP TABLE IF EXISTS vet_visits;
