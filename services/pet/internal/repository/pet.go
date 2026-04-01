package repository

import (
	"context"
	"pet/internal/model"
	"time"

	"github.com/google/uuid"
)

type CreatePetInput struct {
	Pet model.Pet
}

type PetRepository interface {
	Create(ctx context.Context, in CreatePetInput) (*model.Pet, error)
	DeleteByID(ctx context.Context, petID uuid.UUID) error
	GetByID(ctx context.Context, petID uuid.UUID) (*model.Pet, error)
	ListByIDs(ctx context.Context, ids []uuid.UUID, includeArchived bool, offset, limit int) ([]model.Pet, int, error)
	Update(ctx context.Context, petID uuid.UUID, rowVersion int, pet model.Pet) (*model.Pet, error)
	UpdateOwner(ctx context.Context, petID uuid.UUID, rowVersion int, ownerUserID uuid.UUID) (*model.Pet, error)
	UpdatePhoto(ctx context.Context, petID uuid.UUID, rowVersion int, fileID uuid.UUID) (*model.Pet, error)
	UpdateStatus(ctx context.Context, petID uuid.UUID, rowVersion int, status string, missingSince *time.Time, archivedAt *time.Time) (*model.Pet, error)
}
