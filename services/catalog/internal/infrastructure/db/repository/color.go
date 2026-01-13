package repository

import (
	"catalog/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ColorRepo struct {
	db *pgxpool.Pool
}

func NewColorRepo(db *pgxpool.Pool) *ColorRepo {
	return &ColorRepo{db: db}
}

func (r *ColorRepo) List(ctx context.Context, activeOnly bool) ([]model.Color, error) {
	q := `
		SELECT id, name_ru, name_en, hex, is_active, version, created_at, updated_at
		FROM colors
	`
	if activeOnly {
		q += ` WHERE is_active = TRUE`
	}
	q += ` ORDER BY id ASC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Color
	for rows.Next() {
		var c model.Color
		if err := rows.Scan(&c.ID, &c.NameRu, &c.NameEn, &c.Hex, &c.IsActive, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *ColorRepo) GetByID(ctx context.Context, id int) (*model.Color, error) {
	var c model.Color
	err := r.db.QueryRow(ctx, `
		SELECT id, name_ru, name_en, hex, is_active, version, created_at, updated_at
		FROM colors
		WHERE id = $1
	`, id).Scan(&c.ID, &c.NameRu, &c.NameEn, &c.Hex, &c.IsActive, &c.Version, &c.CreatedAt, &c.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *ColorRepo) CreateTx(ctx context.Context, tx pgx.Tx, c *model.Color) error {
	return tx.QueryRow(ctx, `
		INSERT INTO colors (name_ru, name_en, hex, is_active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, c.NameRu, c.NameEn, c.Hex, c.IsActive, c.Version).Scan(&c.ID)
}

func (r *ColorRepo) UpdateTx(ctx context.Context, tx pgx.Tx, c *model.Color) error {
	cmd, err := tx.Exec(ctx, `
		UPDATE colors
		SET name_ru = $2,
		    name_en = $3,
		    is_active = $4,
		    version = $5,
		    updated_at = NOW()
		WHERE id = $1
	`, c.ID, c.NameRu, c.NameEn, c.IsActive, c.Version)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
