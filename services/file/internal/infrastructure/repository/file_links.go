package pgrepo

import (
	"context"

	"file/internal/domain/model"

	"github.com/google/uuid"
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
			id, file_id, usage_type, owner_id, created_at
		) VALUES (
			$1, $2, $3, $4, NOW()
		)
		ON CONFLICT (file_id, usage_type, owner_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query,
		l.ID,
		l.FileID,
		l.UsageType,
		l.OwnerID,
	)
	return err
}

func (r *FileLinkRepository) Delete(ctx context.Context, fileID uuid.UUID, usageType model.FileUsageType, ownerID uuid.UUID) (bool, error) {
	const query = `
		DELETE FROM file_links
		WHERE file_id = $1
		  AND usage_type = $2
		  AND owner_id = $3
	`

	cmd, err := r.db.Exec(ctx, query, fileID, usageType, ownerID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *FileLinkRepository) ListByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error) {
	const query = `
		SELECT id, file_id, usage_type, owner_id, created_at
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
			&l.UsageType,
			&l.OwnerID,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *FileLinkRepository) CountByFileID(ctx context.Context, fileID uuid.UUID) (int64, error) {
	const query = `
		SELECT COUNT(1)
		FROM file_links
		WHERE file_id = $1
	`
	row := r.db.QueryRow(ctx, query, fileID)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
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
