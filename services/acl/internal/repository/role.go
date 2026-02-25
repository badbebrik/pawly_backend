package repository

import (
	"context"

	"github.com/google/uuid"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*RoleView, error)
	ListSystemAndPetRoles(ctx context.Context, petID uuid.UUID) ([]RoleView, error)
	CreateCustom(ctx context.Context, petID uuid.UUID, title string, createdByUserID uuid.UUID) (*RoleView, error)
	DeleteCustomIfUnused(ctx context.Context, petID, roleID uuid.UUID) error
}
