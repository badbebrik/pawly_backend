package repository

import (
	"catalog/internal/model"
	"catalog/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not_found")

type SpeciesRepo struct {
	db *pgxpool.Pool
}

func NewSpeciesRepo(db *pgxpool.Pool) *SpeciesRepo {
	return &SpeciesRepo{db: db}
}

func (r *SpeciesRepo) List(ctx context.Context, activeOnly bool) ([]model.Species, error) {
	q := `
		SELECT id, name_ru, name_en, is_active, version, created_at, updated_at
		FROM species
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

	var out []model.Species
	for rows.Next() {
		var s model.Species
		if err := rows.Scan(&s.ID, &s.NameRu, &s.NameEn, &s.IsActive, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *SpeciesRepo) GetByID(ctx context.Context, id int) (*model.Species, error) {
	var s model.Species
	err := r.db.QueryRow(ctx, `
		SELECT id, name_ru, name_en, is_active, version, created_at, updated_at
		FROM species
		WHERE id = $1
	`, id).Scan(&s.ID, &s.NameRu, &s.NameEn, &s.IsActive, &s.Version, &s.CreatedAt, &s.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *SpeciesRepo) CreateTx(ctx context.Context, tx repository.TxExec, s *model.Species) error {
	row := tx.QueryRow(ctx, `
		INSERT INTO species (name_ru, name_en, is_active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, s.NameRu, s.NameEn, s.IsActive, s.Version)

	if err := row.Scan(&s.ID); err != nil {
		return fmt.Errorf("insert species: %w", err)
	}
	return nil
}

func (r *SpeciesRepo) UpdateTx(ctx context.Context, tx repository.TxExec, s *model.Species) error {
	_, err := tx.Exec(ctx, `
		UPDATE species
		SET name_ru = $2,
		    name_en = $3,
		    is_active = $4,
		    version = $5,
		    updated_at = NOW()
		WHERE id = $1
	`, s.ID, s.NameRu, s.NameEn, s.IsActive, s.Version)

	return err
}
