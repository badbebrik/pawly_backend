package storage

import (
	"context"
	"file/internal/domain/model"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioStorageAdapter struct {
	client *MinioClient
}

func NewMinioStorageAdapter(client *MinioClient) *MinioStorageAdapter {
	return &MinioStorageAdapter{client: client}
}

func (m *MinioStorageAdapter) Bucket() string             { return m.client.Bucket }
func (m *MinioStorageAdapter) DownloadTTL() time.Duration { return m.client.DownloadTTL() }
func (m *MinioStorageAdapter) UploadTTL() time.Duration   { return m.client.UploadTTL() }

func (m *MinioStorageAdapter) PresignPut(ctx context.Context, bucket, objectKey, _ string, expires time.Duration) (string, error) {
	client := m.client.Client
	if m.client.PresignClient != nil {
		client = m.client.PresignClient
	}
	u, err := client.PresignedPutObject(ctx, bucket, objectKey, expires)
	if err != nil {
		return "", fmt.Errorf("minio presign put: %w", err)
	}
	return u.String(), nil
}

func (m *MinioStorageAdapter) PresignGet(ctx context.Context, bucket, objectKey string, expires time.Duration) (string, error) {
	client := m.client.Client
	if m.client.PresignClient != nil {
		client = m.client.PresignClient
	}
	u, err := client.PresignedGetObject(ctx, bucket, objectKey, expires, nil)
	if err != nil {
		return "", fmt.Errorf("minio presign get: %w", err)
	}
	return u.String(), nil
}

func (m *MinioStorageAdapter) StatObject(ctx context.Context, bucket, objectKey string) (model.ObjectInfo, error) {
	info, err := m.client.Client.StatObject(ctx, bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return model.ObjectInfo{}, fmt.Errorf("minio stat object: %w", err)
	}
	return model.ObjectInfo{Size: info.Size, ContentType: info.ContentType}, nil
}

func (m *MinioStorageAdapter) DeleteObject(ctx context.Context, bucket, objectKey string) error {
	if err := m.client.Client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio delete object: %w", err)
	}
	return nil
}
