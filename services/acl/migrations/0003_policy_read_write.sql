-- +goose Up
UPDATE permission_presets
SET policy = jsonb_build_object(
    'pet_read', COALESCE((policy->>'pet_read')::boolean, FALSE),
    'pet_write', COALESCE(
        (policy->>'pet_write')::boolean,
        COALESCE((policy->>'pet_edit')::boolean, FALSE)
            OR COALESCE((policy->>'pet_status_change')::boolean, FALSE)
            OR COALESCE((policy->>'pet_delete')::boolean, FALSE)
    ),
    'log_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'log_write', COALESCE(
        (policy->>'log_write')::boolean,
        COALESCE((policy->>'log_create')::boolean, FALSE)
            OR COALESCE((policy->>'log_edit')::boolean, FALSE)
            OR COALESCE((policy->>'log_delete')::boolean, FALSE)
    ),
    'health_read', COALESCE((policy->>'health_read')::boolean, FALSE),
    'health_write', COALESCE((policy->>'health_write')::boolean, FALSE),
    'task_read', COALESCE((policy->>'task_read')::boolean, FALSE),
    'task_write', COALESCE((policy->>'task_write')::boolean, FALSE),
    'members_read', COALESCE(
        (policy->>'members_read')::boolean,
        COALESCE((policy->>'members_view')::boolean, FALSE)
    ),
    'members_write', COALESCE(
        (policy->>'members_write')::boolean,
        COALESCE((policy->>'members_invite')::boolean, FALSE)
            OR COALESCE((policy->>'members_remove')::boolean, FALSE)
            OR COALESCE((policy->>'members_edit_permissions')::boolean, FALSE)
    )
);

UPDATE pet_memberships
SET policy = jsonb_build_object(
    'pet_read', COALESCE((policy->>'pet_read')::boolean, FALSE),
    'pet_write', COALESCE(
        (policy->>'pet_write')::boolean,
        COALESCE((policy->>'pet_edit')::boolean, FALSE)
            OR COALESCE((policy->>'pet_status_change')::boolean, FALSE)
            OR COALESCE((policy->>'pet_delete')::boolean, FALSE)
    ),
    'log_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'log_write', COALESCE(
        (policy->>'log_write')::boolean,
        COALESCE((policy->>'log_create')::boolean, FALSE)
            OR COALESCE((policy->>'log_edit')::boolean, FALSE)
            OR COALESCE((policy->>'log_delete')::boolean, FALSE)
    ),
    'health_read', COALESCE((policy->>'health_read')::boolean, FALSE),
    'health_write', COALESCE((policy->>'health_write')::boolean, FALSE),
    'task_read', COALESCE((policy->>'task_read')::boolean, FALSE),
    'task_write', COALESCE((policy->>'task_write')::boolean, FALSE),
    'members_read', COALESCE(
        (policy->>'members_read')::boolean,
        COALESCE((policy->>'members_view')::boolean, FALSE)
    ),
    'members_write', COALESCE(
        (policy->>'members_write')::boolean,
        COALESCE((policy->>'members_invite')::boolean, FALSE)
            OR COALESCE((policy->>'members_remove')::boolean, FALSE)
            OR COALESCE((policy->>'members_edit_permissions')::boolean, FALSE)
    )
);

UPDATE pet_invites
SET policy = jsonb_build_object(
    'pet_read', COALESCE((policy->>'pet_read')::boolean, FALSE),
    'pet_write', COALESCE(
        (policy->>'pet_write')::boolean,
        COALESCE((policy->>'pet_edit')::boolean, FALSE)
            OR COALESCE((policy->>'pet_status_change')::boolean, FALSE)
            OR COALESCE((policy->>'pet_delete')::boolean, FALSE)
    ),
    'log_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'log_write', COALESCE(
        (policy->>'log_write')::boolean,
        COALESCE((policy->>'log_create')::boolean, FALSE)
            OR COALESCE((policy->>'log_edit')::boolean, FALSE)
            OR COALESCE((policy->>'log_delete')::boolean, FALSE)
    ),
    'health_read', COALESCE((policy->>'health_read')::boolean, FALSE),
    'health_write', COALESCE((policy->>'health_write')::boolean, FALSE),
    'task_read', COALESCE((policy->>'task_read')::boolean, FALSE),
    'task_write', COALESCE((policy->>'task_write')::boolean, FALSE),
    'members_read', COALESCE(
        (policy->>'members_read')::boolean,
        COALESCE((policy->>'members_view')::boolean, FALSE)
    ),
    'members_write', COALESCE(
        (policy->>'members_write')::boolean,
        COALESCE((policy->>'members_invite')::boolean, FALSE)
            OR COALESCE((policy->>'members_remove')::boolean, FALSE)
            OR COALESCE((policy->>'members_edit_permissions')::boolean, FALSE)
    )
);

-- +goose Down
UPDATE permission_presets
SET policy = jsonb_build_object(
    'pet_read', COALESCE((policy->>'pet_read')::boolean, FALSE),
    'pet_edit', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'pet_status_change', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'pet_delete', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'log_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'log_create', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_edit', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_delete', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_attachments_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'health_read', COALESCE((policy->>'health_read')::boolean, FALSE),
    'health_write', COALESCE((policy->>'health_write')::boolean, FALSE),
    'task_read', COALESCE((policy->>'task_read')::boolean, FALSE),
    'task_write', COALESCE((policy->>'task_write')::boolean, FALSE),
    'members_view', COALESCE((policy->>'members_read')::boolean, FALSE),
    'members_invite', COALESCE((policy->>'members_write')::boolean, FALSE),
    'members_remove', COALESCE((policy->>'members_write')::boolean, FALSE),
    'members_edit_permissions', COALESCE((policy->>'members_write')::boolean, FALSE)
);

UPDATE pet_memberships
SET policy = jsonb_build_object(
    'pet_read', COALESCE((policy->>'pet_read')::boolean, FALSE),
    'pet_edit', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'pet_status_change', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'pet_delete', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'log_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'log_create', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_edit', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_delete', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_attachments_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'health_read', COALESCE((policy->>'health_read')::boolean, FALSE),
    'health_write', COALESCE((policy->>'health_write')::boolean, FALSE),
    'task_read', COALESCE((policy->>'task_read')::boolean, FALSE),
    'task_write', COALESCE((policy->>'task_write')::boolean, FALSE),
    'members_view', COALESCE((policy->>'members_read')::boolean, FALSE),
    'members_invite', COALESCE((policy->>'members_write')::boolean, FALSE),
    'members_remove', COALESCE((policy->>'members_write')::boolean, FALSE),
    'members_edit_permissions', COALESCE((policy->>'members_write')::boolean, FALSE)
);

UPDATE pet_invites
SET policy = jsonb_build_object(
    'pet_read', COALESCE((policy->>'pet_read')::boolean, FALSE),
    'pet_edit', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'pet_status_change', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'pet_delete', COALESCE((policy->>'pet_write')::boolean, FALSE),
    'log_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'log_create', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_edit', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_delete', COALESCE((policy->>'log_write')::boolean, FALSE),
    'log_attachments_read', COALESCE((policy->>'log_read')::boolean, FALSE),
    'health_read', COALESCE((policy->>'health_read')::boolean, FALSE),
    'health_write', COALESCE((policy->>'health_write')::boolean, FALSE),
    'task_read', COALESCE((policy->>'task_read')::boolean, FALSE),
    'task_write', COALESCE((policy->>'task_write')::boolean, FALSE),
    'members_view', COALESCE((policy->>'members_read')::boolean, FALSE),
    'members_invite', COALESCE((policy->>'members_write')::boolean, FALSE),
    'members_remove', COALESCE((policy->>'members_write')::boolean, FALSE),
    'members_edit_permissions', COALESCE((policy->>'members_write')::boolean, FALSE)
);
