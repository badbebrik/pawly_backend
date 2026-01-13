package repository

import (
	"catalog/internal/model"
	"context"
)

type SpeciesRepository interface {
	List(ctx context.Context, activeOnly bool) ([]model.Species, error)
	CreateTx(ctx context.Context, tx TxExec, s *model.Species) error
	UpdateTx(ctx context.Context, tx TxExec, s *model.Species) error
	GetByID(ctx context.Context, id int) (*model.Species, error)
}

type TxExec interface {
	Exec(ctx context.Context, sql string, args ...any) (any, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
}
