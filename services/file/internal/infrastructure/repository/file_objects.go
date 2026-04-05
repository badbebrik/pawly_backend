package pgrepo

import (
	"context"
	"errors"
	"file/internal/model"
	"file/internal/repository"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileObjectRepository struct {
	db *pgxpool.Pool
}

func NewFileObjectRepository(db *pgxpool.Pool) *FileObjectRepository {
	return &FileObjectRepository{db: db}
}

func (r *FileObjectRepository) Create(ctx context.Context, f *model.FileObject) error {
	const query = `
		INSERT INTO files (
			id, storage_bucket, storage_key, status, mime_type,
			size_bytes, original_filename, created_at, updated_at, upload_expires_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, NOW(), NOW(), $8, NULL
		)
	`
	_, err := r.db.Exec(ctx, query,
		f.ID,
		f.StorageBucket,
		f.StorageKey,
		f.Status,
		f.MimeType,
		f.SizeBytes,
		f.OriginalFilename,
		f.UploadExpiresAt,
	)
	return err
}

func (r *FileObjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FileObject, error) {
	const query = `
		SELECT id, storage_bucket, storage_key, status, mime_type,
		       size_bytes, original_filename, created_at, updated_at, upload_expires_at, deleted_at
		FROM files
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var f model.FileObject
	err := row.Scan(
		&f.ID,
		&f.StorageBucket,
		&f.StorageKey,
		&f.Status,
		&f.MimeType,
		&f.SizeBytes,
		&f.OriginalFilename,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.UploadExpiresAt,
		&f.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &f, nil
}

func (r *FileObjectRepository) ListExpiredUploading(ctx context.Context, now time.Time, limit int) ([]model.FileObject, error) {
	const query = `
		SELECT id, storage_bucket, storage_key, status, mime_type,
		       size_bytes, original_filename, created_at, updated_at, upload_expires_at, deleted_at
		FROM files
		WHERE status = $1
		  AND upload_expires_at < $2
		ORDER BY upload_expires_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, model.FileStatusUploading, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FileObject, 0, limit)
	for rows.Next() {
		var f model.FileObject
		if err := rows.Scan(
			&f.ID,
			&f.StorageBucket,
			&f.StorageKey,
			&f.Status,
			&f.MimeType,
			&f.SizeBytes,
			&f.OriginalFilename,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.UploadExpiresAt,
			&f.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, f)
	}

	return items, rows.Err()
}

func (r *FileObjectRepository) ListPendingDelete(ctx context.Context, limit int) ([]model.FileObject, error) {
	const query = `
		SELECT id, storage_bucket, storage_key, status, mime_type,
		       size_bytes, original_filename, created_at, updated_at, upload_expires_at, deleted_at
		FROM files
		WHERE status = $1
		ORDER BY updated_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, model.FileStatusPendingDelete, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FileObject, 0, limit)
	for rows.Next() {
		var f model.FileObject
		if err := rows.Scan(
			&f.ID,
			&f.StorageBucket,
			&f.StorageKey,
			&f.Status,
			&f.MimeType,
			&f.SizeBytes,
			&f.OriginalFilename,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.UploadExpiresAt,
			&f.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, f)
	}

	return items, rows.Err()
}

func (r *FileObjectRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus) error {
	const query = `
		UPDATE files
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`

	cmd, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *FileObjectRepository) ConfirmUpload(ctx context.Context, id uuid.UUID, sizeBytes int64) error {
	const query = `
		UPDATE files
		SET size_bytes = $2,
		    status = $3,
		    updated_at = NOW()
		WHERE id = $1
	`

	cmd, err := r.db.Exec(ctx, query, id, sizeBytes, model.FileStatusReady)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *FileObjectRepository) MarkPendingDelete(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE files
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, query, id, model.FileStatusPendingDelete)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *FileObjectRepository) MarkDeleted(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE files
		SET status = $2,
		    deleted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, query, id, model.FileStatusDeleted)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *FileObjectRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM files
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
