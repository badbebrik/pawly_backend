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

type FileObjectRepository struct {
	db *pgxpool.Pool
}

func NewFileObjectRepository(db *pgxpool.Pool) *FileObjectRepository {
	return &FileObjectRepository{db: db}
}

func (r *FileObjectRepository) Create(ctx context.Context, f *model.FileObject) error {
	const query = `
		INSERT INTO file_objects (
			id, bucket, object_key, status, mime_type,
			size_bytes, created_by_user_id, created_at, updated_at, upload_expires_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, NOW(), NOW(), $8
		)
	`
	_, err := r.db.Exec(ctx, query,
		f.ID,
		f.Bucket,
		f.ObjectKey,
		f.Status,
		f.MimeType,
		f.SizeBytes,
		f.CreatedByUserID,
		f.UploadExpiresAt,
	)
	return err
}

func (r *FileObjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FileObject, error) {
	const query = `
		SELECT id, bucket, object_key, status, mime_type,
		       size_bytes, created_by_user_id, created_at, updated_at, upload_expires_at
		FROM file_objects
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var f model.FileObject
	err := row.Scan(
		&f.ID,
		&f.Bucket,
		&f.ObjectKey,
		&f.Status,
		&f.MimeType,
		&f.SizeBytes,
		&f.CreatedByUserID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.UploadExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &f, nil
}

func (r *FileObjectRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus) error {
	const query = `
		UPDATE file_objects
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
		UPDATE file_objects
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

func (r *FileObjectRepository) GetByObjectKey(ctx context.Context, bucket string, objectKey string) (*model.FileObject, error) {
	const query = `
		SELECT id, bucket, object_key, status, mime_type,
		       size_bytes, created_by_user_id, created_at, updated_at, upload_expires_at
		FROM file_objects
		WHERE bucket = $1 AND object_key = $2
	`

	row := r.db.QueryRow(ctx, query, bucket, objectKey)

	var f model.FileObject
	err := row.Scan(
		&f.ID,
		&f.Bucket,
		&f.ObjectKey,
		&f.Status,
		&f.MimeType,
		&f.SizeBytes,
		&f.CreatedByUserID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.UploadExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &f, nil
}

func (r *FileObjectRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE file_objects
		SET status = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, query, id, model.FileStatusFailed)
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
		DELETE FROM file_objects
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
