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
    scheduled_for TIMESTAMPTZ NOT NULL,
    recurrence_rule TEXT NULL CHECK (
        recurrence_rule IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY')
    ),
    recurrence_interval INT NULL,
    recurrence_until TIMESTAMPTZ NULL,
    status TEXT NOT NULL CHECK (
        status IN ('ACTIVE', 'DONE', 'CANCELLED')
    ),
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
    CONSTRAINT scheduled_items_recurrence_invariant CHECK (
        (recurrence_rule IS NULL AND recurrence_interval IS NULL AND recurrence_until IS NULL)
        OR
        (recurrence_rule IS NOT NULL AND recurrence_interval IS NOT NULL)
    )
);

CREATE INDEX idx_scheduled_items_pet_scheduled_active ON scheduled_items (pet_id, scheduled_for ASC, id ASC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_scheduled_items_pet_status_scheduled_active ON scheduled_items (pet_id, status, scheduled_for ASC, id ASC)
WHERE deleted_at IS NULL;

CREATE INDEX idx_scheduled_items_source_active ON scheduled_items (source_type, source_id)
WHERE deleted_at IS NULL AND source_id IS NOT NULL;

CREATE UNIQUE INDEX uq_scheduled_items_health_source_active
ON scheduled_items (source_type, source_id)
WHERE deleted_at IS NULL
  AND source_id IS NOT NULL
  AND source_type IN ('VET_VISIT', 'VACCINATION', 'PROCEDURE');

CREATE TABLE scheduled_item_push_dispatches (
    id UUID PRIMARY KEY,
    scheduled_item_id UUID NOT NULL REFERENCES scheduled_items(id) ON DELETE CASCADE,
    dispatch_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (scheduled_item_id, dispatch_key)
);

CREATE INDEX idx_scheduled_item_push_dispatches_item_created ON scheduled_item_push_dispatches (scheduled_item_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS scheduled_item_push_dispatches;
DROP TABLE IF EXISTS scheduled_items;
