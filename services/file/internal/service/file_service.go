package service

import (
	"context"
	"file/internal/model"
	"file/internal/repository"
	"file/internal/storage"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	PresignPut(ctx context.Context, bucket, objectKey, contentType string, expires time.Duration) (string, error)
	PresignGet(ctx context.Context, bucket, objectKey string, expires time.Duration) (string, error)
	UploadTTL() time.Duration
	DownloadTTL() time.Duration
}

type FileService struct {
	objects repository.FileObjectRepository
	links   repository.FileLinkRepository
	storage Storage
}

func NewFileService(objects repository.FileObjectRepository, links repository.FileLinkRepository, storage Storage) *FileService {
	return &FileService{
		objects: objects,
		links:   links,
		storage: storage,
	}
}

type InitUploadParams struct {
	MimeType          string
	ExpectedSizeBytes *int64
	CreatedByUserID   uuid.UUID
}

type UploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
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
		ID:              fileID,
		Bucket:          "",
		ObjectKey:       objectKey,
		Status:          model.FileStatusUploading,
		MimeType:        p.MimeType,
		SizeBytes:       p.ExpectedSizeBytes,
		CreatedByUserID: p.CreatedByUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
		UploadExpiresAt: now.Add(s.storage.UploadTTL()),
	}

	if st, ok := s.storage.(*storage.MinioStorageAdapter); ok {
		file.Bucket = st.Bucket()
	}

	if err := s.objects.Create(ctx, file); err != nil {
		return nil, nil, err
	}

	url, err := s.storage.PresignPut(ctx, file.Bucket, file.ObjectKey, file.MimeType, s.storage.UploadTTL())
	if err != nil {
		return nil, nil, err
	}

	upload := &UploadInfo{
		Method:    "PUT",
		URL:       url,
		Headers:   map[string]string{"Content-Type": file.MimeType},
		ExpiresAt: now.Add(s.storage.UploadTTL()),
	}

	return file, upload, nil
}

type ConfirmUploadParams struct {
	FileID    uuid.UUID
	SizeBytes int64
}

func (s *FileService) ConfirmUpload(ctx context.Context, p ConfirmUploadParams) (*model.FileObject, error) {
	if p.SizeBytes <= 0 {
		return nil, ErrInvalidInput
	}

	f, err := s.objects.GetByID(ctx, p.FileID)
	if err != nil {
		return nil, err
	}

	if f.Status != model.FileStatusUploading {
		return nil, ErrInvalidState
	}

	if time.Now().UTC().After(f.UploadExpiresAt) {
		return nil, ErrUploadExpired
	}

	if err := s.objects.ConfirmUpload(ctx, p.FileID, p.SizeBytes); err != nil {
		return nil, err
	}

	f.Status = model.FileStatusReady
	f.SizeBytes = &p.SizeBytes
	f.UpdatedAt = time.Now().UTC()

	return f, nil
}

func (s *FileService) GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error) {
	f, err := s.objects.GetByID(ctx, fileID)
	if err != nil {
		return "", time.Time{}, err
	}
	if f.Status != model.FileStatusReady {
		return "", time.Time{}, ErrNotReady
	}

	ttl := s.storage.DownloadTTL()
	url, err := s.storage.PresignGet(ctx, f.Bucket, f.ObjectKey, ttl)
	if err != nil {
		return "", time.Time{}, err
	}

	return url, time.Now().UTC().Add(ttl), nil
}

type LinkParams struct {
	FileID        uuid.UUID
	OwnerService  model.OwnerService
	OwnerType     string
	OwnerID       uuid.UUID
	PetID         *uuid.UUID
	CreatedByUser uuid.UUID
}

func (s *FileService) Link(ctx context.Context, p LinkParams) (*model.FileLink, error) {
	f, err := s.objects.GetByID(ctx, p.FileID)
	if err != nil {
		return nil, err
	}

	if f.Status != model.FileStatusReady {
		return nil, ErrNotReady
	}

	link := &model.FileLink{
		ID:              uuid.New(),
		FileID:          p.FileID,
		OwnerService:    p.OwnerService,
		OwnerType:       p.OwnerType,
		OwnerID:         p.OwnerID,
		PetID:           p.PetID,
		CreatedByUserID: p.CreatedByUser,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.links.Create(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *FileService) Unlink(ctx context.Context, fileID uuid.UUID, ownerService model.OwnerService, ownerType string, ownerID uuid.UUID) (bool, error) {
	return s.links.Delete(ctx, fileID, ownerService, ownerType, ownerID)
}

func (s *FileService) GetFile(ctx context.Context, id uuid.UUID) (*model.FileObject, error) {
	return s.objects.GetByID(ctx, id)
}

func (s *FileService) ListLinksByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error) {
	return s.links.ListByFileID(ctx, fileID)
}
