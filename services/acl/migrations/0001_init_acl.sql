-- +goose Up
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('SYSTEM', 'CUSTOM')),
    pet_id UUID NULL,
    code TEXT NULL,
    title TEXT NOT NULL,
    created_by_user_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT roles_kind_invariant CHECK (
        (kind = 'SYSTEM' AND pet_id IS NULL AND code IS NOT NULL AND created_by_user_id IS NULL)
        OR
        (kind = 'CUSTOM' AND pet_id IS NOT NULL AND code IS NULL AND created_by_user_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX roles_system_code ON roles (code) WHERE kind = 'SYSTEM';
CREATE UNIQUE INDEX roles_custom_pet_title ON roles (pet_id, title) WHERE kind = 'CUSTOM';
CREATE INDEX idx_roles_pet_id ON roles (pet_id);
CREATE INDEX idx_roles_kind ON roles (kind);

CREATE TABLE permission_presets (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    role_code TEXT NULL,
    policy JSONB NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_presets_role_code ON permission_presets (role_code);

CREATE TABLE pet_memberships (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'REMOVED')),
    role_id UUID NOT NULL REFERENCES roles(id),
    policy JSONB NOT NULL,
    base_preset_id UUID NULL REFERENCES permission_presets(id),
    is_primary_owner BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    removed_at TIMESTAMPTZ NULL,
    removed_by_user_id UUID NULL,

    CONSTRAINT memberships_removed_fields CHECK (
        (status = 'REMOVED' AND removed_at IS NOT NULL AND removed_by_user_id IS NOT NULL)
        OR
        (status = 'ACTIVE' AND removed_at IS NULL AND removed_by_user_id IS NULL)
    ),
    CONSTRAINT memberships_primary_owner_active CHECK (
        is_primary_owner = FALSE OR status = 'ACTIVE'
    )
);

CREATE UNIQUE INDEX memberships_pet_user ON pet_memberships (pet_id, user_id);
CREATE UNIQUE INDEX memberships_primary_owner ON pet_memberships (pet_id) WHERE is_primary_owner = TRUE;
CREATE INDEX memberships_user_status ON pet_memberships (user_id, status);
CREATE INDEX memberships_pet_status ON pet_memberships (pet_id, status);
CREATE INDEX memberships_role_id ON pet_memberships (role_id);

CREATE TABLE pet_invites (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    created_by_user_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'CONSUMED', 'EXPIRED', 'REVOKED')),
    token_hash TEXT NOT NULL,
    code TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    consumed_by_user_id UUID NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    policy JSONB NOT NULL,
    base_preset_id UUID NULL REFERENCES permission_presets(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT invites_code_format CHECK (code ~ '^[0-9]{6}$'),
    CONSTRAINT invites_consumed_fields CHECK (
        (status = 'CONSUMED' AND consumed_at IS NOT NULL AND consumed_by_user_id IS NOT NULL)
        OR
        (status <> 'CONSUMED' AND consumed_at IS NULL AND consumed_by_user_id IS NULL)
    )
);

CREATE UNIQUE INDEX invites_token_hash ON pet_invites (token_hash);
CREATE UNIQUE INDEX invites_active_code ON pet_invites (code) WHERE status = 'ACTIVE';
CREATE INDEX invites_pet_status ON pet_invites (pet_id, status);
CREATE INDEX invites_expires_at ON pet_invites (expires_at);
CREATE INDEX invites_created_by ON pet_invites (created_by_user_id);

-- +goose Down
DROP TABLE IF EXISTS pet_invites;
DROP TABLE IF EXISTS pet_memberships;
DROP TABLE IF EXISTS permission_presets;
DROP TABLE IF EXISTS roles;
