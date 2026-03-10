-- +goose Up
CREATE TABLE log_types (
    id UUID PRIMARY KEY,
    scope TEXT NOT NULL CHECK (scope IN ('SYSTEM', 'CUSTOM')),
    pet_id UUID NULL,
    code TEXT NULL,
    name TEXT NOT NULL,
    metric_requirements JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NULL,
    row_version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL,
    CONSTRAINT log_types_scope_invariant CHECK (
        (scope = 'SYSTEM' AND pet_id IS NULL)
        OR
        (scope = 'CUSTOM' AND pet_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_log_types_system_code_active ON log_types (code)
WHERE scope = 'SYSTEM' AND deleted_at IS NULL AND code IS NOT NULL;

CREATE UNIQUE INDEX uq_log_types_custom_pet_name_active ON log_types (pet_id, lower(name))
WHERE scope = 'CUSTOM' AND deleted_at IS NULL;

CREATE INDEX idx_log_types_pet_active ON log_types (pet_id, deleted_at);

CREATE TABLE metrics (
    id UUID PRIMARY KEY,
    scope TEXT NOT NULL CHECK (scope IN ('SYSTEM', 'CUSTOM')),
    pet_id UUID NULL,
    code TEXT NULL,
    name TEXT NOT NULL,
    input_kind TEXT NOT NULL CHECK (input_kind IN ('NUMERIC', 'SCALE')),
    unit_code TEXT NULL,
    min_value NUMERIC NULL,
    max_value NUMERIC NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NULL,
    row_version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL,
    CONSTRAINT metrics_scope_invariant CHECK (
        (scope = 'SYSTEM' AND pet_id IS NULL)
        OR
        (scope = 'CUSTOM' AND pet_id IS NOT NULL)
    ),
    CONSTRAINT metrics_range_invariant CHECK (
        (min_value IS NULL AND max_value IS NULL)
        OR
        (min_value IS NOT NULL AND max_value IS NOT NULL AND min_value <= max_value)
    )
);

CREATE UNIQUE INDEX uq_metrics_system_code_active ON metrics (code)
WHERE scope = 'SYSTEM' AND deleted_at IS NULL AND code IS NOT NULL;

CREATE UNIQUE INDEX uq_metrics_custom_pet_name_active ON metrics (pet_id, lower(name))
WHERE scope = 'CUSTOM' AND deleted_at IS NULL;

CREATE INDEX idx_metrics_pet_active ON metrics (pet_id, deleted_at);

CREATE TABLE logs (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    log_type_id UUID NULL REFERENCES log_types(id),
    description TEXT NULL,
    source TEXT NOT NULL CHECK (source IN ('USER', 'HEALTH')),
    source_entity_type TEXT NULL,
    source_entity_id UUID NULL,
    row_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL
);

CREATE INDEX idx_logs_pet_occurred_id_active ON logs (pet_id, occurred_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_logs_pet_source_active ON logs (pet_id, source)
WHERE deleted_at IS NULL;

CREATE TABLE metric_values (
    id UUID PRIMARY KEY,
    log_id UUID NOT NULL REFERENCES logs(id) ON DELETE CASCADE,
    metric_id UUID NOT NULL REFERENCES metrics(id),
    value_num NUMERIC NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (log_id, metric_id)
);

CREATE INDEX idx_metric_values_metric_id ON metric_values (metric_id);

CREATE TABLE log_attachment_refs (
    id UUID PRIMARY KEY,
    log_id UUID NOT NULL REFERENCES logs(id) ON DELETE CASCADE,
    file_id UUID NOT NULL,
    file_type TEXT NOT NULL CHECK (file_type IN ('image', 'pdf', 'other')),
    added_by_user_id UUID NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (log_id, file_id)
);

CREATE INDEX idx_log_attachment_refs_file_id ON log_attachment_refs (file_id);

INSERT INTO metrics (
    id, scope, pet_id, code, name, input_kind, unit_code, min_value, max_value, created_at, updated_at
) VALUES
    ('10000000-0000-0000-0000-000000000001', 'SYSTEM', NULL, 'WEIGHT', 'Вес', 'NUMERIC', 'kg', 0.1, 200, NOW(), NOW()),
    ('10000000-0000-0000-0000-000000000002', 'SYSTEM', NULL, 'TEMPERATURE', 'Температура', 'NUMERIC', 'c', 30, 45, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO log_types (
    id, scope, pet_id, code, name, metric_requirements, created_at, updated_at
) VALUES
    (
        '20000000-0000-0000-0000-000000000001',
        'SYSTEM',
        NULL,
        'WEIGHING',
        'Взвешивание',
        '[{"metric_id":"10000000-0000-0000-0000-000000000001","is_required":true}]'::jsonb,
        NOW(),
        NOW()
    ),
    (
        '20000000-0000-0000-0000-000000000002',
        'SYSTEM',
        NULL,
        'TEMPERATURE',
        'Температура',
        '[{"metric_id":"10000000-0000-0000-0000-000000000002","is_required":true}]'::jsonb,
        NOW(),
        NOW()
    )
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS log_attachment_refs;
DROP TABLE IF EXISTS metric_values;
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS log_types;
