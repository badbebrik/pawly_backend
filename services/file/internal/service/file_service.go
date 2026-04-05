package service

import (
	"context"
	"file/internal/model"
	"file/internal/repository"
	"file/internal/storage"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Storage interface {
	PresignPut(ctx context.Context, bucket, objectKey, contentType string, expires time.Duration) (string, error)
	PresignGet(ctx context.Context, bucket, objectKey string, expires time.Duration) (string, error)
	StatObject(ctx context.Context, bucket, objectKey string) (model.ObjectInfo, error)
	DeleteObject(ctx context.Context, bucket, objectKey string) error
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
	OriginalFilename  *string
}

type UploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

type CleanupResult struct {
	DeletedPending      int
	MarkedPendingDelete int
	DeletedExpired      int
}

func (s *FileService) InitUpload(ctx context.Context, p InitUploadParams) (*model.FileObject, *UploadInfo, error) {
	if p.MimeType == "" {
		return nil, nil, ErrInvalidInput
	}
	if p.OriginalFilename != nil {
		trimmed := strings.TrimSpace(*p.OriginalFilename)
		if trimmed == "" {
			p.OriginalFilename = nil
		} else {
			p.OriginalFilename = &trimmed
		}
	}

	fileID := uuid.New()
	objectKey := fileID.String()
	now := time.Now().UTC()

	file := &model.FileObject{
		ID:               fileID,
		StorageBucket:    "",
		StorageKey:       objectKey,
		Status:           model.FileStatusUploading,
		MimeType:         p.MimeType,
		SizeBytes:        p.ExpectedSizeBytes,
		OriginalFilename: p.OriginalFilename,
		CreatedAt:        now,
		UpdatedAt:        now,
		UploadExpiresAt:  now.Add(s.storage.UploadTTL()),
	}

	if st, ok := s.storage.(*storage.MinioStorageAdapter); ok {
		file.StorageBucket = st.Bucket()
	}

	if err := s.objects.Create(ctx, file); err != nil {
		return nil, nil, err
	}

	url, err := s.storage.PresignPut(ctx, file.StorageBucket, file.StorageKey, file.MimeType, s.storage.UploadTTL())
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
	f, err := s.objects.GetByID(ctx, p.FileID)
	if err != nil {
		return nil, err
	}

	if f.Status != model.FileStatusUploading {
		return nil, ErrInvalidState
	}

	if f.SizeBytes != nil && *f.SizeBytes != p.SizeBytes {
		return nil, ErrInvalidInput
	}

	if time.Now().UTC().After(f.UploadExpiresAt) {
		return nil, ErrUploadExpired
	}

	info, err := s.storage.StatObject(ctx, f.StorageBucket, f.StorageKey)
	if err != nil {
		return nil, err
	}
	if info.Size <= 0 {
		return nil, ErrInvalidState
	}
	if p.SizeBytes > 0 && p.SizeBytes != info.Size {
		return nil, ErrInvalidInput
	}

	if err := s.objects.ConfirmUpload(ctx, p.FileID, info.Size); err != nil {
		return nil, err
	}

	f.Status = model.FileStatusReady
	f.SizeBytes = &info.Size
	if info.ContentType != "" {
		f.MimeType = info.ContentType
	}
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
	url, err := s.storage.PresignGet(ctx, f.StorageBucket, f.StorageKey, ttl)
	if err != nil {
		return "", time.Time{}, err
	}

	return url, time.Now().UTC().Add(ttl), nil
}

type LinkParams struct {
	FileID    uuid.UUID
	UsageType model.FileUsageType
	OwnerID   uuid.UUID
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
		ID:        uuid.New(),
		FileID:    p.FileID,
		UsageType: p.UsageType,
		OwnerID:   p.OwnerID,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.links.Create(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *FileService) Unlink(ctx context.Context, fileID uuid.UUID, usageType model.FileUsageType, ownerID uuid.UUID) (bool, error) {
	return s.links.Delete(ctx, fileID, usageType, ownerID)
}

func (s *FileService) GetFile(ctx context.Context, id uuid.UUID) (*model.FileObject, error) {
	return s.objects.GetByID(ctx, id)
}

func (s *FileService) ListLinksByFileID(ctx context.Context, fileID uuid.UUID) ([]model.FileLink, error) {
	return s.links.ListByFileID(ctx, fileID)
}

func (s *FileService) DeleteFileIfUnlinked(ctx context.Context, fileID uuid.UUID) (*model.FileObject, error) {
	f, err := s.objects.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	count, err := s.links.CountByFileID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrHasLinks
	}

	switch f.Status {
	case model.FileStatusDeleted:
		return f, nil
	case model.FileStatusReady, model.FileStatusPendingDelete:
	default:
		return nil, ErrInvalidState
	}

	now := time.Now().UTC()
	if err := s.storage.DeleteObject(ctx, f.StorageBucket, f.StorageKey); err != nil {
		if err := s.objects.MarkPendingDelete(ctx, fileID); err != nil {
			return nil, err
		}
		f.Status = model.FileStatusPendingDelete
		f.UpdatedAt = now
		return f, nil
	}

	if err := s.objects.MarkDeleted(ctx, fileID); err != nil {
		return nil, err
	}

	f.Status = model.FileStatusDeleted
	f.UpdatedAt = now
	f.DeletedAt = &now
	return f, nil
}

func (s *FileService) RunCleanupBatch(ctx context.Context, limit int) (CleanupResult, error) {
	if limit <= 0 {
		limit = 100
	}

	var result CleanupResult

	expired, err := s.objects.ListExpiredUploading(ctx, time.Now().UTC(), limit)
	if err != nil {
		return result, err
	}
	for i := range expired {
		f := expired[i]
		_ = s.storage.DeleteObject(ctx, f.StorageBucket, f.StorageKey)
		if err := s.objects.DeleteByID(ctx, f.ID); err != nil {
			continue
		}
		result.DeletedExpired++
	}

	pending, err := s.objects.ListPendingDelete(ctx, limit)
	if err != nil {
		return result, err
	}
	for i := range pending {
		f := pending[i]

		count, err := s.links.CountByFileID(ctx, f.ID)
		if err != nil || count > 0 {
			continue
		}

		if err := s.storage.DeleteObject(ctx, f.StorageBucket, f.StorageKey); err != nil {
			result.MarkedPendingDelete++
			continue
		}

		if err := s.objects.MarkDeleted(ctx, f.ID); err != nil {
			continue
		}
		result.DeletedPending++
	}

	return result, nil
}
