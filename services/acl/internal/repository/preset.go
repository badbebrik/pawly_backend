package repository

import (
	"acl/internal/model"
	"context"

	"github.com/google/uuid"
)

type PermissionPresetView struct {
	ID       uuid.UUID
	Name     string
	RoleCode string
	Policy   model.Policy
}

type PresetRepository interface {
	ListSystem(ctx context.Context) ([]PermissionPresetView, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
