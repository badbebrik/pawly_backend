package repository

import (
	"catalog/internal/model"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SpeciesRepository interface {
	List(ctx context.Context, activeOnly bool) ([]model.Species, error)
	CreateTx(ctx context.Context, tx pgx.Tx, s *model.Species) error
	UpdateTx(ctx context.Context, tx pgx.Tx, s *model.Species) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Species, error)
}
