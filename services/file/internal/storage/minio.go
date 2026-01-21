package storage

import (
	"context"
	"file/internal/config"
	"fmt"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
	Bucket string

	uploadTTL   time.Duration
	downloadTTL time.Duration
}

func NewMinio(cfg *config.Config) (*MinioClient, error) {
	useSSL, err := strconv.ParseBool(cfg.MinioUseSSL)
	if err != nil {
		return nil, fmt.Errorf("minio use ssl: %w", err)
	}

	mc, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	uploadTTL, err := parseSeconds(cfg.UploadURLTTLSeconds, "upload ttl")
	if err != nil {
		return nil, err
	}

	downloadTTL, err := parseSeconds(cfg.DownloadURLTTLSeconds, "download tll")
	if err != nil {
		return nil, err
	}

	return &MinioClient{
		Client:      mc,
		Bucket:      cfg.MinioBucket,
		uploadTTL:   uploadTTL,
		downloadTTL: downloadTTL,
	}, nil
}

func (m *MinioClient) UploadTTL() time.Duration {
	return m.uploadTTL
}

func (m *MinioClient) DownloadTTL() time.Duration {
	return m.downloadTTL
}

func parseSeconds(raw string, label string) (time.Duration, error) {
	if raw == "" {
		return 0, fmt.Errorf("%s: empty", label)
	}

	sec, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", label, err)
	}

	if sec <= 0 {
		return 0, fmt.Errorf("%s: must be > 0", label)
	}
	return time.Duration(sec) * time.Second, nil
}

func (m *MinioClient) EnsureBucket(ctx context.Context) error {
	exists, err := m.Client.BucketExists(ctx, m.Bucket)
	if err != nil {
		return fmt.Errorf("minio bucket exists: %w", err)
	}

	if exists {
		return nil
	}

	if err := m.Client.MakeBucket(ctx, m.Bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("minio make bucket: %w", err)
	}

	return nil
}
