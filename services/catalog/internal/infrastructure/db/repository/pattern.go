package repository

import (
	"catalog/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PatternRepo struct {
	db *pgxpool.Pool
}

func NewPatternRepo(db *pgxpool.Pool) *PatternRepo {
	return &PatternRepo{db: db}
}

func (r *PatternRepo) List(ctx context.Context, activeOnly bool) ([]model.Pattern, error) {
	q := `
		SELECT id, name_ru, name_en, icon_key, is_active, version, created_at, updated_at
		FROM patterns
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

	var out []model.Pattern
	for rows.Next() {
		var p model.Pattern
		if err := rows.Scan(&p.ID, &p.NameRu, &p.NameEn, &p.IconKey, &p.IsActive, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PatternRepo) GetByID(ctx context.Context, id int) (*model.Pattern, error) {
	var p model.Pattern
	err := r.db.QueryRow(ctx, `
		SELECT id, name_ru, name_en, icon_key, is_active, version, created_at, updated_at
		FROM patterns
		WHERE id = $1
	`, id).Scan(&p.ID, &p.NameRu, &p.NameEn, &p.IconKey, &p.IsActive, &p.Version, &p.CreatedAt, &p.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *PatternRepo) CreateTx(ctx context.Context, tx pgx.Tx, p *model.Pattern) error {
	return tx.QueryRow(ctx, `
		INSERT INTO patterns (name_ru, name_en, is_active, icon_key, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, p.NameRu, p.NameEn, p.IconKey, p.IsActive, p.Version).Scan(&p.ID)
}

func (r *PatternRepo) UpdateTx(ctx context.Context, tx pgx.Tx, p *model.Pattern) error {
	cmd, err := tx.Exec(ctx, `
		UPDATE patterns
		SET name_ru = $2,
		    name_en = $3,
		    icon_key = $4,
		    is_active = $5,
		    version = $6,
		    updated_at = NOW()
		WHERE id = $1
	`, p.ID, p.NameRu, p.NameEn, p.IconKey, p.IsActive, p.Version)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
