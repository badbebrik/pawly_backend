-- +goose Up
CREATE TABLE files (
    id                UUID PRIMARY KEY,
    storage_bucket    TEXT        NOT NULL,
    storage_key       TEXT        NOT NULL,
    status            TEXT        NOT NULL CHECK (status IN ('UPLOADING', 'READY', 'PENDING_DELETE', 'DELETED')),
    mime_type         TEXT        NOT NULL,
    size_bytes        BIGINT      NULL,
    original_filename TEXT        NULL,
    upload_expires_at TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ NULL,
    UNIQUE (storage_bucket, storage_key)
);

CREATE TABLE file_links (
    id         UUID PRIMARY KEY,
    file_id    UUID        NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    usage_type TEXT        NOT NULL CHECK (usage_type IN (
        'USER_AVATAR',
        'PET_AVATAR',
        'LOG_ATTACHMENT',
        'HEALTH_ATTACHMENT',
        'CHAT_ATTACHMENT'
    )),
    owner_id   UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (file_id, usage_type, owner_id)
);

CREATE INDEX idx_file_links_file_id ON file_links (file_id);
CREATE INDEX idx_file_links_usage_owner ON file_links (usage_type, owner_id);

-- +goose Down
DROP TABLE IF EXISTS file_links;
DROP TABLE IF EXISTS files;
