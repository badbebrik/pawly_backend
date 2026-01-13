package repository

import (
	"catalog/internal/model"
	"context"

	"github.com/jackc/pgx/v5"
)

type PatternRepository interface {
	List(ctx context.Context, activeOnly bool) ([]model.Pattern, error)
	GetByID(ctx context.Context, id int) (*model.Pattern, error)
	CreateTx(ctx context.Context, tx pgx.Tx, p *model.Pattern) error
	UpdateTx(ctx context.Context, tx pgx.Tx, p *model.Pattern) error
}
