CREATE TABLE profiles (
    user_id       UUID PRIMARY KEY,
    first_name    TEXT,
    last_name     TEXT,
    avatar_url    TEXT,
    phone         TEXT,
    locale        VARCHAR(8)  NOT NULL DEFAULT 'ru',
    timezone      VARCHAR(64) NOT NULL DEFAULT 'Europe/Moscow',
    date_format   VARCHAR(32) NOT NULL DEFAULT 'dd.MM.yyyy',
    notifications JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
