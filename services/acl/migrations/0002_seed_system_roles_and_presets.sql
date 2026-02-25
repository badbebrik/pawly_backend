-- +goose Up
INSERT INTO roles (id, kind, pet_id, code, title, created_by_user_id, created_at, updated_at)
VALUES
    ('a1111111-1111-1111-1111-111111111111', 'SYSTEM', NULL, 'OWNER',     'Owner',      NULL, NOW(), NOW()),
    ('a2222222-2222-2222-2222-222222222222', 'SYSTEM', NULL, 'CO_OWNER',  'Co-owner',   NULL, NOW(), NOW()),
    ('a3333333-3333-3333-3333-333333333333', 'SYSTEM', NULL, 'VET',       'Veterinary', NULL, NOW(), NOW()),
    ('a4444444-4444-4444-4444-444444444444', 'SYSTEM', NULL, 'PETSITTER', 'Petsitter',  NULL, NOW(), NOW()),
    ('a5555555-5555-5555-5555-555555555555', 'SYSTEM', NULL, 'WALKER',    'Walker',     NULL, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO permission_presets (id, name, role_code, policy, is_system, created_at, updated_at)
VALUES
    (
      'b1111111-1111-1111-1111-111111111111',
      'Owner Full Access',
      'OWNER',
      '{
        "pet_read": true,
        "pet_edit": true,
        "pet_status_change": true,
        "pet_delete": true,
        "log_read": true,
        "log_create": true,
        "log_edit": true,
        "log_delete": true,
        "log_attachments_read": true,
        "health_read": true,
        "health_write": true,
        "task_read": true,
        "task_write": true,
        "members_view": true,
        "members_invite": true,
        "members_remove": true,
        "members_edit_permissions": true
      }'::jsonb,
      TRUE,
      NOW(),
      NOW()
    ),
    (
      'b2222222-2222-2222-2222-222222222222',
      'Co-owner Standard',
      'CO_OWNER',
      '{
        "pet_read": true,
        "pet_edit": true,
        "pet_status_change": true,
        "pet_delete": false,
        "log_read": true,
        "log_create": true,
        "log_edit": true,
        "log_delete": false,
        "log_attachments_read": true,
        "health_read": true,
        "health_write": true,
        "task_read": true,
        "task_write": true,
        "members_view": true,
        "members_invite": true,
        "members_remove": false,
        "members_edit_permissions": false
      }'::jsonb,
      TRUE,
      NOW(),
      NOW()
    ),
    (
      'b3333333-3333-3333-3333-333333333333',
      'Vet Access',
      'VET',
      '{
        "pet_read": true,
        "pet_edit": false,
        "pet_status_change": false,
        "pet_delete": false,
        "log_read": true,
        "log_create": true,
        "log_edit": true,
        "log_delete": false,
        "log_attachments_read": true,
        "health_read": true,
        "health_write": true,
        "task_read": true,
        "task_write": false,
        "members_view": false,
        "members_invite": false,
        "members_remove": false,
        "members_edit_permissions": false
      }'::jsonb,
      TRUE,
      NOW(),
      NOW()
    ),
    (
      'b4444444-4444-4444-4444-444444444444',
      'Petsitter Access',
      'PETSITTER',
      '{
        "pet_read": true,
        "pet_edit": false,
        "pet_status_change": false,
        "pet_delete": false,
        "log_read": true,
        "log_create": true,
        "log_edit": false,
        "log_delete": false,
        "log_attachments_read": true,
        "health_read": true,
        "health_write": false,
        "task_read": true,
        "task_write": true,
        "members_view": false,
        "members_invite": false,
        "members_remove": false,
        "members_edit_permissions": false
      }'::jsonb,
      TRUE,
      NOW(),
      NOW()
    ),
    (
      'b5555555-5555-5555-5555-555555555555',
      'Walker Access',
      'WALKER',
      '{
        "pet_read": true,
        "pet_edit": false,
        "pet_status_change": false,
        "pet_delete": false,
        "log_read": true,
        "log_create": true,
        "log_edit": false,
        "log_delete": false,
        "log_attachments_read": true,
        "health_read": false,
        "health_write": false,
        "task_read": true,
        "task_write": true,
        "members_view": false,
        "members_invite": false,
        "members_remove": false,
        "members_edit_permissions": false
      }'::jsonb,
      TRUE,
      NOW(),
      NOW()
    )
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM permission_presets
WHERE id IN (
    'b1111111-1111-1111-1111-111111111111',
    'b2222222-2222-2222-2222-222222222222',
    'b3333333-3333-3333-3333-333333333333',
    'b4444444-4444-4444-4444-444444444444',
    'b5555555-5555-5555-5555-555555555555'
);

DELETE FROM roles
WHERE id IN (
    'a1111111-1111-1111-1111-111111111111',
    'a2222222-2222-2222-2222-222222222222',
    'a3333333-3333-3333-3333-333333333333',
    'a4444444-4444-4444-4444-444444444444',
    'a5555555-5555-5555-5555-555555555555'
);
