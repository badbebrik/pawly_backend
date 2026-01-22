-- +goose Up
CREATE TABLE profiles (
    user_id       UUID PRIMARY KEY,
    first_name    TEXT,
    last_name     TEXT,
    avatar_file_id UUID NULL,
    public_contact_settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    extra_contacts JSONB NOT NULL DEFAULT '{}'::jsonb,
    phone         TEXT,
    locale        VARCHAR(8)  NOT NULL DEFAULT 'ru',
    timezone      VARCHAR(64) NOT NULL DEFAULT 'Europe/Moscow',
    date_format   VARCHAR(32) NOT NULL DEFAULT 'dd.MM.yyyy',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS profiles;
