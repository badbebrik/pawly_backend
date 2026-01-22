package service

import (
	"context"
	"file/internal/model"
	"file/internal/repository"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	PresignPut(ctx context.Context, bucket, objectKey, contentType string, expires time.Duration) (string, error)
	PresignGet(ctx context.Context, bucket, objectKey, expires time.Duration) (string, error)
	UploadTTL() time.Duration
	DownloadTTL() time.Duration
}

type FileService struct {
	objects repository.FileObjectRepository
	links repository.FileLinkRepository
	storage Storage
}

func NewFileService(objects repository.FileObjectRepository, links repository.FileLinkRepository, storage Storage) *FileService {
	return &FileService{
		objects: objects,
		links: links,
		storage: storage,
	}
}

type InitUploadParams struct {
	MimeType string
	ExpectedSizeBytes *int64
	CreatedByUserID uuid.UUID
}

type UploadInfo struct {
	Method string
	URL string
	Headers map[string]string
	ExpiresAt time.Time
}

func (s *FileService) InitUpload(ctx context.Context, p InitUploadParams) (*model.FileObject, *UploadInfo, error) {
	if p.MimeType == "" {
		return nil, nil, ErrInvalidInput
	}

	fileID := uuid.New()
	objectKey := fileID.String()
	now := time.Now().UTC()

	file := &model.FileObject{
		ID: fileID,
		Bucket: "",
		ObjectKey: objectKey,
		Status: model.FileStatusUploading,
		MimeType: p.MimeType,
		SizeBytes: p.ExpectedSizeBytes,
		CreatedByUserID: p.CreatedByUserID
		CreatedAt: now,
		UpdatedAt: now,
		UploadExpiresAt: now.Add(s.storage.UploadTTL()),
	}

	
}