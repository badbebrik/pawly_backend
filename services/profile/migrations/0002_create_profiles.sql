-- +goose Up
CREATE TABLE profiles (
    user_id       UUID PRIMARY KEY,
    first_name    TEXT,
    last_name     TEXT,
    avatar_file_id UUID NULL,
    locale        VARCHAR(8)  NOT NULL DEFAULT 'ru',
    timezone      VARCHAR(64) NOT NULL DEFAULT 'UTC',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS profiles;
