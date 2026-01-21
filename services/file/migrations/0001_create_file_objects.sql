CREATE TABLE file_objects (
    id               UUID PRIMARY KEY,
    bucket           TEXT        NOT NULL,
    object_key       TEXT        NOT NULL,
    status           TEXT        NOT NULL CHECK (status IN ('UPLOADING', 'READY', 'FAILED', 'DELETED')),
    mime_type        TEXT        NOT NULL,
    size_bytes       BIGINT      NULL,
    created_by_user_id UUID      NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    upload_expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE (bucket, object_key)
);

CREATE TABLE file_links (
    id                UUID PRIMARY KEY,
    file_id           UUID        NOT NULL REFERENCES file_objects(id),
    owner_service     TEXT        NOT NULL CHECK (owner_service IN ('PROFILE', 'PET', 'GUIDE', 'LOG', 'HEALTH', 'CHAT', 'CATALOG')),
    owner_type        TEXT        NOT NULL,
    owner_id          UUID        NOT NULL,
    pet_id            UUID        NULL,
    created_by_user_id UUID       NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
