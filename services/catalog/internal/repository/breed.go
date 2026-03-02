package repository

import (
	"catalog/internal/model"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BreedRepository interface {
	List(ctx context.Context, speciesID *uuid.UUID, activeOnly bool) ([]model.Breed, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Breed, error)
	CreateTx(ctx context.Context, tx pgx.Tx, b *model.Breed) error
	UpdateTx(ctx context.Context, tx pgx.Tx, b *model.Breed) error
}
