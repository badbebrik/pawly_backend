package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

type MinioStorageAdapter struct {
	client *MinioClient
}

func NewMinioStorageAdapter(client *MinioClient) *MinioStorageAdapter {
	return &MinioStorageAdapter{client: client}
}

func (m *MinioStorageAdapter) Bucket() string {
	return m.client.Bucket
}

func (m *MinioStorageAdapter) DownloadTTL() time.Duration {
	return m.client.DownloadTTL()
}

func (m *MinioStorageAdapter) UploadTTL() time.Duration {
	return m.client.UploadTTL()
}

func (m *MinioStorageAdapter) PresignPut(ctx context.Context, bucket, objectKey, contentType string, expires time.Duration) (string, error) {
	params := make(url.Values)
	params.Set("response-content-type", contentType)
	u, err := m.client.Client.PresignedPutObject(ctx, bucket, objectKey, expires)
	if err != nil {
		return "", fmt.Errorf("minio presign put: %w", err)
	}
	return u.String(), nil
}

func (m *MinioStorageAdapter) PresignGet(ctx context.Context, bucket, objectKey string, expires time.Duration) (string, error) {
	u, err := m.client.Client.PresignedGetObject(ctx, bucket, objectKey, expires, nil)
	if err != nil {
		return "", fmt.Errorf("minio presign get: %w", err)
	}
	return u.String(), nil
}
