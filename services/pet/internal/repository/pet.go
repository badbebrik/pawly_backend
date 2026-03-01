package repository

import (
	"context"
	"pet/internal/model"

	"github.com/google/uuid"
)

type CreatePetInput struct {
	Pet model.Pet
}

type PetRepository interface {
	Create(ctx context.Context, in CreatePetInput) (*model.Pet, error)
	GetByID(ctx context.Context, petID uuid.UUID) (*model.Pet, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID, includeArchived bool, offset, limit int) ([]model.Pet, int, error)
}
