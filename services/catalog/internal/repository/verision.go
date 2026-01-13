package repository

import "context"

type VersionRepository interface {
	Get(ctx context.Context) (int, error)
	BumpTx(ctx context.Context, tx Tx) (newVersion int, err error)
}

type Tx interface {
	QueryRow(ctx context.Context, sql string, args ...any) Row
}

type Row interface {
	Scan(dest ...any) error
}
