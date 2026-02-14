package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

type FileClient interface {
	InitUpload(ctx context.Context, mimeType string, expectedSize int64, userID uuid.UUID) (uuid.UUID, UploadInfo, error)
	ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) error
	GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error)
	LinkAvatar(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) error
}
