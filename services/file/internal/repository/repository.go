package repository

import (
	"context"
	"file/internal/model"
	"time"

	"github.com/google/uuid"
)

type FileObjectRepository interface {
	Create(ctx context.Context, f *model.FileObject) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.FileObject, error)
	ListExpiredUploading(ctx context.Context, now time.Time, limit int) ([]model.FileObject, error)
	ListPendingDelete(ctx context.Context, limit int) ([]model.FileObject, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus) error
	ConfirmUpload(ctx context.Context, id uuid.UUID, sizeBytes int64) error
	MarkPendingDelete(ctx context.Context, id uuid.UUID) error
	MarkDeleted(ctx context.Context, id uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type FileLinkRepository interface {
	Create(ctx context.Context, l *model.FileLink) error
	ListByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error)
	CountByFileID(ctx context.Context, fileID uuid.UUID) (int64, error)
	Delete(ctx context.Context, fileID uuid.UUID, usageType model.FileUsageType, ownerID uuid.UUID) (bool, error)
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) (int64, error)
}
