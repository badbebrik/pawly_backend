package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type VersionRepository interface {
	Get(ctx context.Context) (int, error)
	BumpTx(ctx context.Context, tx pgx.Tx) (newVersion int, err error)
}
