package repository

import (
	"context"

	"github.com/google/uuid"
)

type RoleRepository interface {
	ListSystemAndPetRoles(ctx context.Context, petID uuid.UUID) ([]RoleView, error)
}
