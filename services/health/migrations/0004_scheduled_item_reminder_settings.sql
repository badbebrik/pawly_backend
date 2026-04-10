-- +goose Up
ALTER TABLE scheduled_items
    ADD COLUMN push_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN remind_offset_minutes INT NULL;

UPDATE scheduled_items
SET
    push_enabled = FALSE,
    remind_offset_minutes = NULL
WHERE source_type IN ('VET_VISIT', 'VACCINATION', 'PROCEDURE');

ALTER TABLE scheduled_items
    ADD CONSTRAINT scheduled_items_remind_offset_minutes_check
    CHECK (remind_offset_minutes IS NULL OR remind_offset_minutes >= 0);

CREATE INDEX idx_scheduled_items_push_enabled_active
ON scheduled_items (push_enabled, starts_at ASC, id ASC)
WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_scheduled_items_push_enabled_active;

ALTER TABLE scheduled_items
    DROP CONSTRAINT IF EXISTS scheduled_items_remind_offset_minutes_check;

ALTER TABLE scheduled_items
    DROP COLUMN IF EXISTS remind_offset_minutes,
    DROP COLUMN IF EXISTS push_enabled;
