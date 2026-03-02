package repository

import (
	"catalog/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BreedRepo struct {
	db *pgxpool.Pool
}

func NewBreedRepo(db *pgxpool.Pool) *BreedRepo {
	return &BreedRepo{db: db}
}

func (r *BreedRepo) List(ctx context.Context, speciesID *uuid.UUID, activeOnly bool) ([]model.Breed, error) {
	q := `
		SELECT id, species_id, name_ru, name_en, is_active, version, created_at, updated_at
		FROM breeds
	`

	args := make([]any, 0, 2)
	where := ""

	if speciesID != nil {
		args = append(args, *speciesID)
		where = fmt.Sprintf("species_id = $%d", len(args))
	}

	if activeOnly {
		if where != "" {
			where += " AND "
		}
		where += "is_active = TRUE"
	}

	if where != "" {
		q += " WHERE " + where
	}
	q += ` ORDER BY id ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Breed, 0)
	for rows.Next() {
		var b model.Breed
		if err := rows.Scan(
			&b.ID,
			&b.SpeciesID,
			&b.NameRu,
			&b.NameEn,
			&b.IsActive,
			&b.Version,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BreedRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Breed, error) {
	var b model.Breed
	err := r.db.QueryRow(ctx, `
		SELECT id, species_id, name_ru, name_en, is_active, version, created_at, updated_at
		FROM breeds
		WHERE id = $1
	`, id).Scan(
		&b.ID,
		&b.SpeciesID,
		&b.NameRu,
		&b.NameEn,
		&b.IsActive,
		&b.Version,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &b, err
}

func (r *BreedRepo) CreateTx(ctx context.Context, tx pgx.Tx, b *model.Breed) error {
	return tx.QueryRow(ctx, `
		INSERT INTO breeds (species_id, name_ru, name_en, is_active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, b.SpeciesID, b.NameRu, b.NameEn, b.IsActive, b.Version).Scan(&b.ID)
}

func (r *BreedRepo) UpdateTx(ctx context.Context, tx pgx.Tx, b *model.Breed) error {
	cmd, err := tx.Exec(ctx, `
		UPDATE breeds
		SET species_id = $2,
		    name_ru = $3,
		    name_en = $4,
		    is_active = $5,
		    version = $6,
		    updated_at = NOW()
		WHERE id = $1
	`, b.ID, b.SpeciesID, b.NameRu, b.NameEn, b.IsActive, b.Version)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
