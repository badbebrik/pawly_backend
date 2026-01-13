package repository

import (
	"catalog/internal/repository"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VersionRepo struct {
	db *pgxpool.Pool
}

func NewVersionRepo(db *pgxpool.Pool) *VersionRepo {
	return &VersionRepo{db: db}
}

func (r *VersionRepo) Get(ctx context.Context) (int, error) {
	var v int
	err := r.db.QueryRow(ctx, `SELECT version FROM catalog_version WHERE id = 1`).Scan(&v)
	return v, err
}

func (r *VersionRepo) BumpTx(ctx context.Context, tx repository.Tx) (int, error) {
	var v int
	err := tx.QueryRow(ctx, `
		UPDATE catalog_version
		SET version = version + 1
		WHERE id = 1
		RETURNING version
	`).Scan(&v)
	if err != nil {
		return 0, fmt.Errorf("bump version: %w", err)
	}
	return v, nil
}
