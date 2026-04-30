-- +goose Up
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('SYSTEM', 'CUSTOM')),
    pet_id UUID NULL,
    code TEXT NULL,
    title TEXT NOT NULL,
    policy JSONB NOT NULL,
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

CREATE TABLE pet_memberships (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'REMOVED')),
    role_id UUID NOT NULL,
    policy JSONB NOT NULL,
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
    token TEXT NOT NULL,
    code TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    consumed_by_user_id UUID NULL,
    role_id UUID NOT NULL,
    policy JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT invites_code_format CHECK (code ~ '^[0-9]{6}$'),
    CONSTRAINT invites_consumed_fields CHECK (
        (status = 'CONSUMED' AND consumed_at IS NOT NULL AND consumed_by_user_id IS NOT NULL)
        OR
        (status <> 'CONSUMED' AND consumed_at IS NULL AND consumed_by_user_id IS NULL)
    )
);

CREATE UNIQUE INDEX invites_token ON pet_invites (token);
CREATE UNIQUE INDEX invites_active_code ON pet_invites (code) WHERE status = 'ACTIVE';
CREATE INDEX invites_pet_status ON pet_invites (pet_id, status);
CREATE INDEX invites_expires_at ON pet_invites (expires_at);
CREATE INDEX invites_created_by ON pet_invites (created_by_user_id);

INSERT INTO roles (id, kind, pet_id, code, title, policy, created_by_user_id, created_at, updated_at)
VALUES
    (
      'a1111111-1111-1111-1111-111111111111',
      'SYSTEM',
      NULL,
      'OWNER',
      'Owner',
      '{
        "pet_read": true,
        "pet_write": true,
        "log_read": true,
        "log_write": true,
        "health_read": true,
        "health_write": true,
        "members_read": true,
        "members_write": true
      }'::jsonb,
      NULL,
      NOW(),
      NOW()
    ),
    (
      'a2222222-2222-2222-2222-222222222222',
      'SYSTEM',
      NULL,
      'CO_OWNER',
      'Co-owner',
      '{
        "pet_read": true,
        "pet_write": true,
        "log_read": true,
        "log_write": true,
        "health_read": true,
        "health_write": true,
        "members_read": true,
        "members_write": true
      }'::jsonb,
      NULL,
      NOW(),
      NOW()
    ),
    (
      'a3333333-3333-3333-3333-333333333333',
      'SYSTEM',
      NULL,
      'VET',
      'Veterinary',
      '{
        "pet_read": true,
        "pet_write": false,
        "log_read": true,
        "log_write": true,
        "health_read": true,
        "health_write": true,
        "members_read": false,
        "members_write": false
      }'::jsonb,
      NULL,
      NOW(),
      NOW()
    ),
    (
      'a4444444-4444-4444-4444-444444444444',
      'SYSTEM',
      NULL,
      'PETSITTER',
      'Petsitter',
      '{
        "pet_read": true,
        "pet_write": false,
        "log_read": true,
        "log_write": true,
        "health_read": true,
        "health_write": false,
        "members_read": false,
        "members_write": false
      }'::jsonb,
      NULL,
      NOW(),
      NOW()
    ),
    (
      'a5555555-5555-5555-5555-555555555555',
      'SYSTEM',
      NULL,
      'WALKER',
      'Walker',
      '{
        "pet_read": true,
        "pet_write": false,
        "log_read": true,
        "log_write": true,
        "health_read": false,
        "health_write": false,
        "members_read": false,
        "members_write": false
      }'::jsonb,
      NULL,
      NOW(),
      NOW()
    )
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS pet_invites;
DROP TABLE IF EXISTS pet_memberships;
DROP TABLE IF EXISTS roles;
