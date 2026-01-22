package pgrepo

import (
	"context"
	"errors"
	"file/internal/model"
	"file/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileLinkRepository struct {
	db *pgxpool.Pool
}

func NewFileLinkRepository(db *pgxpool.Pool) *FileLinkRepository {
	return &FileLinkRepository{db: db}
}

func (r *FileLinkRepository) Create(ctx context.Context, l *model.FileLink) error {
	const query = `
		INSERT INTO file_links (
			id, file_id, owner_service, owner_type, owner_id,
			pet_id, created_by_user_id, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, NOW()
		)
	`

	_, err := r.db.Exec(ctx, query,
		l.ID,
		l.FileID,
		l.OwnerService,
		l.OwnerType,
		l.OwnerID,
		l.PetID,
		l.CreatedByUserID,
	)
	return err
}

func (r *FileLinkRepository) Delete(ctx context.Context, fileID uuid.UUID, ownerService model.OwnerService, ownerType string, ownerID uuid.UUID) (bool, error) {
	const query = `
		DELETE FROM file_links
		WHERE file_id = $1
		  AND owner_service = $2
		  AND owner_type = $3
		  AND owner_id = $4
	`

	cmd, err := r.db.Exec(ctx, query, fileID, ownerService, ownerType, ownerID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *FileLinkRepository) ListByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error) {
	const query = `
		SELECT id, file_id, owner_service, owner_type, owner_id,
		       pet_id, created_by_user_id, created_at
		FROM file_links
		WHERE file_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FileLink, 0)
	for rows.Next() {
		var l model.FileLink
		if err := rows.Scan(
			&l.ID,
			&l.FileID,
			&l.OwnerService,
			&l.OwnerType,
			&l.OwnerID,
			&l.PetID,
			&l.CreatedByUserID,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *FileLinkRepository) DeleteByPetID(ctx context.Context, petID uuid.UUID) error {
	const query = `
		DELETE FROM file_links
		WHERE pet_id = $1
	`

	_, err := r.db.Exec(ctx, query, petID)
	return err
}

func (r *FileLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FileLink, error) {
	const query = `
		SELECT id, file_id, owner_service, owner_type, owner_id,
		       pet_id, created_by_user_id, created_at
		FROM file_links
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var l model.FileLink
	err := row.Scan(
		&l.ID,
		&l.FileID,
		&l.OwnerService,
		&l.OwnerType,
		&l.OwnerID,
		&l.PetID,
		&l.CreatedByUserID,
		&l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &l, nil
}

func (r *FileLinkRepository) ListByOwner(ctx context.Context, ownerService model.OwnerService, ownerType string, ownerID uuid.UUID) ([]model.FileLink, error) {
	const query = `
		SELECT id, file_id, owner_service, owner_type, owner_id,
		       pet_id, created_by_user_id, created_at
		FROM file_links
		WHERE owner_service = $1
		  AND owner_type = $2
		  AND owner_id = $3
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, ownerService, ownerType, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FileLink, 0)
	for rows.Next() {
		var l model.FileLink
		if err := rows.Scan(
			&l.ID,
			&l.FileID,
			&l.OwnerService,
			&l.OwnerType,
			&l.OwnerID,
			&l.PetID,
			&l.CreatedByUserID,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *FileLinkRepository) ListByPetID(ctx context.Context, petID uuid.UUID) ([]model.FileLink, error) {
	const query = `
		SELECT id, file_id, owner_service, owner_type, owner_id,
		       pet_id, created_by_user_id, created_at
		FROM file_links
		WHERE pet_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FileLink, 0)
	for rows.Next() {
		var l model.FileLink
		if err := rows.Scan(
			&l.ID,
			&l.FileID,
			&l.OwnerService,
			&l.OwnerType,
			&l.OwnerID,
			&l.PetID,
			&l.CreatedByUserID,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *FileLinkRepository) DeleteByFileID(ctx context.Context, fileID uuid.UUID) (int64, error) {
	const query = `
		DELETE FROM file_links
		WHERE file_id = $1
	`
	cmd, err := r.db.Exec(ctx, query, fileID)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

func (r *FileLinkRepository) Exists(ctx context.Context, fileID uuid.UUID, ownerService model.OwnerService, ownerType string, ownerID uuid.UUID) (bool, error) {
	const query = `
		SELECT 1
		FROM file_links
		WHERE file_id = $1
		  AND owner_service = $2
		  AND owner_type = $3
		  AND owner_id = $4
		LIMIT 1
	`
	row := r.db.QueryRow(ctx, query, fileID, ownerService, ownerType, ownerID)
	var x int
	if err := row.Scan(&x); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
