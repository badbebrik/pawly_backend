-- +goose Up
CREATE TABLE vet_visits (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'COMPLETED')),
    visit_type TEXT NOT NULL CHECK (visit_type IN ('CHECKUP', 'SYMPTOM', 'FOLLOW_UP', 'VACCINATION', 'PROCEDURE', 'OTHER')),
    title TEXT NULL,
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

CREATE TABLE health_dictionary_items (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('PROCEDURE_TYPE', 'MEDICAL_RECORD_TYPE', 'VACCINATION_TARGET')),
    pet_id UUID NULL,
    code TEXT NULL,
    name TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NULL,
    CONSTRAINT health_dictionary_items_scope_invariant CHECK (
        (is_system = TRUE AND pet_id IS NULL AND code IS NOT NULL)
        OR
        (is_system = FALSE AND pet_id IS NOT NULL AND code IS NULL)
    )
);

CREATE UNIQUE INDEX uq_health_dictionary_system_code
ON health_dictionary_items (kind, code)
WHERE is_system = TRUE;

CREATE UNIQUE INDEX uq_health_dictionary_custom_pet_name_active
ON health_dictionary_items (kind, pet_id, lower(name))
WHERE is_system = FALSE AND is_archived = FALSE;

CREATE INDEX idx_health_dictionary_pet_kind_active
ON health_dictionary_items (pet_id, kind, is_archived, name);

CREATE TABLE vaccinations (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    generated_from_id UUID NULL REFERENCES vaccinations(id),
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'COMPLETED')),
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
CREATE UNIQUE INDEX uq_vaccinations_generated_from_active ON vaccinations (generated_from_id)
WHERE generated_from_id IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE vaccination_target_links (
    vaccination_id UUID NOT NULL REFERENCES vaccinations(id),
    target_item_id UUID NOT NULL REFERENCES health_dictionary_items(id),
    pet_id UUID NOT NULL,
    created_by_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (vaccination_id, target_item_id)
);

CREATE INDEX idx_vaccination_target_links_target
ON vaccination_target_links (pet_id, target_item_id, vaccination_id);

CREATE TABLE procedures (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    generated_from_id UUID NULL REFERENCES procedures(id),
    status TEXT NOT NULL CHECK (status IN ('PLANNED', 'COMPLETED')),
    procedure_type_item_id UUID NOT NULL REFERENCES health_dictionary_items(id),
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
CREATE UNIQUE INDEX uq_procedures_generated_from_active ON procedures (generated_from_id)
WHERE generated_from_id IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE medical_records (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    record_type_item_id UUID NOT NULL REFERENCES health_dictionary_items(id),
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

INSERT INTO health_dictionary_items (id, kind, pet_id, code, name, is_system, created_at, updated_at) VALUES
    ('31000000-0000-0000-0000-000000000001', 'PROCEDURE_TYPE', NULL, 'PARASITE_TREATMENT', 'Обработка от эктопаразитов', TRUE, NOW(), NOW()),
    ('31000000-0000-0000-0000-000000000002', 'PROCEDURE_TYPE', NULL, 'DEWORMING', 'Дегельминтизация', TRUE, NOW(), NOW()),
    ('31000000-0000-0000-0000-000000000003', 'PROCEDURE_TYPE', NULL, 'HYGIENE', 'Гигиена', TRUE, NOW(), NOW()),
    ('31000000-0000-0000-0000-000000000004', 'PROCEDURE_TYPE', NULL, 'WOUND_CARE', 'Обработка раны', TRUE, NOW(), NOW()),
    ('31000000-0000-0000-0000-000000000005', 'PROCEDURE_TYPE', NULL, 'GROOMING', 'Груминг', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000001', 'MEDICAL_RECORD_TYPE', NULL, 'DIAGNOSIS', 'Диагноз', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000002', 'MEDICAL_RECORD_TYPE', NULL, 'ALLERGY', 'Аллергия', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000003', 'MEDICAL_RECORD_TYPE', NULL, 'CHRONIC_CONDITION', 'Хроническое состояние', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000004', 'MEDICAL_RECORD_TYPE', NULL, 'INJURY', 'Травма', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000005', 'MEDICAL_RECORD_TYPE', NULL, 'SURGERY', 'Операция', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000006', 'MEDICAL_RECORD_TYPE', NULL, 'CLINICAL_NOTE', 'Клиническая заметка', TRUE, NOW(), NOW()),
    ('32000000-0000-0000-0000-000000000007', 'MEDICAL_RECORD_TYPE', NULL, 'LAB_TEST', 'Анализ', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000001', 'VACCINATION_TARGET', NULL, 'RABIES', 'Бешенство', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000002', 'VACCINATION_TARGET', NULL, 'LEPTOSPIROSIS', 'Лептоспироз', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000003', 'VACCINATION_TARGET', NULL, 'CANINE_DISTEMPER', 'Чума плотоядных', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000004', 'VACCINATION_TARGET', NULL, 'BORDETELLA', 'Бордетеллез', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000005', 'VACCINATION_TARGET', NULL, 'FELV', 'Вирус лейкоза', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000006', 'VACCINATION_TARGET', NULL, 'CANINE_ADENOVIRUS', 'Аденовирус', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000007', 'VACCINATION_TARGET', NULL, 'CANINE_PARVOVIRUS', 'Парвовирус', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000008', 'VACCINATION_TARGET', NULL, 'CANINE_PARAINFLUENZA', 'Парагрипп', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000009', 'VACCINATION_TARGET', NULL, 'FELINE_PANLEUKOPENIA', 'Панлейкопения', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000010', 'VACCINATION_TARGET', NULL, 'FELINE_CALICIVIRUS', 'Калицивироз', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000011', 'VACCINATION_TARGET', NULL, 'FELINE_HERPESVIRUS', 'Герпесвирус', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000012', 'VACCINATION_TARGET', NULL, 'MYXOMATOSIS', 'Миксоматоз', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000013', 'VACCINATION_TARGET', NULL, 'RABBIT_HEMORRHAGIC_DISEASE', 'Вирусная геморрагическая болезнь', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000014', 'VACCINATION_TARGET', NULL, 'BORRELIA', 'Боррелиоз', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000015', 'VACCINATION_TARGET', NULL, 'CANINE_INFLUENZA', 'Грипп', TRUE, NOW(), NOW()),
    ('33000000-0000-0000-0000-000000000016', 'VACCINATION_TARGET', NULL, 'CHLAMYDIA_FELIS', 'Хламидиоз', TRUE, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS medical_records;
DROP TABLE IF EXISTS procedures;
DROP TABLE IF EXISTS vaccination_target_links;
DROP TABLE IF EXISTS vaccinations;
DROP TABLE IF EXISTS health_dictionary_items;
DROP TABLE IF EXISTS entity_relations;
DROP TABLE IF EXISTS vet_visits;
