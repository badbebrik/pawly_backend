-- +goose Up
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    device_id TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('IOS', 'ANDROID')),
    push_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id),
    UNIQUE (platform, push_token)
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens (user_id);

CREATE TABLE pet_push_settings (
    user_id UUID NOT NULL,
    pet_id UUID NOT NULL,
    scheduled_items_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, pet_id)
);

CREATE INDEX idx_pet_push_settings_pet_id ON pet_push_settings (pet_id);

-- +goose Down
DROP TABLE IF EXISTS pet_push_settings;
DROP TABLE IF EXISTS device_tokens;
