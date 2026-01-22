package repository

import (
	"context"
	"file/internal/model"

	"github.com/google/uuid"
)

type FileObjectRepository interface {
	Create(ctx context.Context, f *model.FileObject) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.FileObject, error)
	GetByObjectKey(ctx context.Context, bucket string, objectKey string) (*model.FileObject, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.FileStatus) error
	ConfirmUpload(ctx context.Context, id uuid.UUID, sizeBytes int64) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type FileLinkRepository interface {
	Create(ctx context.Context, l *model.FileLink) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.FileLink, error)
	ListByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error)
	ListByOwner(ctx context.Context, ownerService model.OwnerService, ownerType string, ownerID uuid.UUID) ([]model.FileLink, error)
	ListByPetID(ctx context.Context, petID uuid.UUID) ([]model.FileLink, error)

	Delete(ctx context.Context, fileID uuid.UUID, ownerService model.OwnerService, ownerType string, ownerID uuid.UUID) (bool, error)
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) (int64, error)
	DeleteByPetID(ctx context.Context, petID uuid.UUID) error
}
