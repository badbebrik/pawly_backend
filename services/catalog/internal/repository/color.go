package repository

import (
	"catalog/internal/model"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ColorRepository interface {
	List(ctx context.Context, activeOnly bool) ([]model.Color, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Color, error)
	CreateTx(ctx context.Context, tx pgx.Tx, c *model.Color) error
	UpdateTx(ctx context.Context, tx pgx.Tx, c *model.Color) error
}
