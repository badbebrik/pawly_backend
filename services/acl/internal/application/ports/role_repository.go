package ports

import (
	"acl/internal/domain/model"
	"context"

	"github.com/google/uuid"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*RoleView, error)
	ListSystemAndPetRoles(ctx context.Context, petID uuid.UUID) ([]RoleView, error)
	CreateCustom(ctx context.Context, petID uuid.UUID, title string, policy model.Policy, createdByUserID uuid.UUID) (*RoleView, error)
	UpdateCustom(ctx context.Context, petID, roleID uuid.UUID, title string, policy model.Policy) (*RoleView, error)
	DeleteCustomIfUnused(ctx context.Context, petID, roleID uuid.UUID) error
}
