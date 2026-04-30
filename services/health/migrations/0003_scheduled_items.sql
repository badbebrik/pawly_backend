-- +goose Up
CREATE TABLE scheduled_items (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    source_type TEXT NOT NULL CHECK (
        source_type IN ('MANUAL', 'LOG_TYPE', 'PET_EVENT', 'VET_VISIT', 'VACCINATION', 'PROCEDURE')
    ),
    source_id UUID NULL,
    title TEXT NOT NULL,
    note TEXT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    push_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    remind_offset_minutes INT NULL,
    recurrence_rule TEXT NULL CHECK (
        recurrence_rule IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY')
    ),
    recurrence_interval INT NULL,
    recurrence_until TIMESTAMPTZ NULL,
    occurrences_generated_until TIMESTAMPTZ NULL,
    row_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by_user_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id UUID NULL,
    CONSTRAINT scheduled_items_recurrence_interval_check CHECK (
        recurrence_interval IS NULL OR recurrence_interval > 0
    ),
    CONSTRAINT scheduled_items_remind_offset_minutes_check CHECK (
        remind_offset_minutes IS NULL OR remind_offset_minutes >= 0
    ),
    CONSTRAINT scheduled_items_recurrence_invariant CHECK (
        (recurrence_rule IS NULL AND recurrence_interval IS NULL AND recurrence_until IS NULL)
        OR
        (recurrence_rule IS NOT NULL AND recurrence_interval IS NOT NULL)
    ),
    CONSTRAINT scheduled_items_recurrence_until_check CHECK (
        recurrence_until IS NULL OR recurrence_until >= starts_at
    )
);

CREATE INDEX idx_scheduled_items_pet_starts_active ON scheduled_items (pet_id, starts_at ASC, id ASC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_scheduled_items_source_active ON scheduled_items (source_type, source_id)
WHERE deleted_at IS NULL AND source_id IS NOT NULL;

CREATE UNIQUE INDEX uq_scheduled_items_health_source_active
ON scheduled_items (source_type, source_id)
WHERE deleted_at IS NULL
  AND source_id IS NOT NULL
  AND source_type IN ('VET_VISIT', 'VACCINATION', 'PROCEDURE');

CREATE INDEX idx_scheduled_items_push_enabled_active
ON scheduled_items (push_enabled, starts_at ASC, id ASC)
WHERE deleted_at IS NULL;

CREATE TABLE scheduled_item_occurrences (
    id UUID PRIMARY KEY,
    scheduled_item_id UUID NOT NULL REFERENCES scheduled_items(id),
    pet_id UUID NOT NULL,
    scheduled_for TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (scheduled_item_id, scheduled_for)
);

CREATE INDEX idx_scheduled_item_occurrences_pet_day
ON scheduled_item_occurrences (pet_id, scheduled_for ASC, id ASC);

CREATE INDEX idx_scheduled_item_occurrences_item_scheduled
ON scheduled_item_occurrences (scheduled_item_id, scheduled_for ASC, id ASC);

CREATE TABLE scheduled_item_push_dispatches (
    id UUID PRIMARY KEY,
    scheduled_item_occurrence_id UUID NOT NULL REFERENCES scheduled_item_occurrences(id),
    dispatch_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (scheduled_item_occurrence_id, dispatch_key)
);

CREATE INDEX idx_scheduled_item_push_dispatches_occurrence_created ON scheduled_item_push_dispatches (scheduled_item_occurrence_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS scheduled_item_push_dispatches;
DROP TABLE IF EXISTS scheduled_item_occurrences;
DROP TABLE IF EXISTS scheduled_items;
