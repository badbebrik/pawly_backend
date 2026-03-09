package storage

import (
	"context"
	"file/internal/config"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client        *minio.Client
	PresignClient *minio.Client
	Bucket        string

	uploadTTL   time.Duration
	downloadTTL time.Duration
}

func NewMinio(cfg *config.Config) (*MinioClient, error) {
	useSSL, err := strconv.ParseBool(cfg.MinioUseSSL)
	if err != nil {
		return nil, fmt.Errorf("minio use ssl: %w", err)
	}

	bucketLookup, err := parseBucketLookup(cfg.MinioBucketLookup)
	if err != nil {
		return nil, err
	}

	minioEndpoint, minioSecure, err := normalizeEndpoint(cfg.MinioEndpoint, useSSL, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("minio endpoint: %w", err)
	}

	mc, err := minio.New(minioEndpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure:       minioSecure,
		Region:       cfg.MinioRegion,
		BucketLookup: bucketLookup,
	})

	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	presignEndpoint := minioEndpoint
	presignSecure := minioSecure
	if cfg.MinioPublicEndpoint != "" {
		presignEndpoint, presignSecure, err = normalizeEndpoint(cfg.MinioPublicEndpoint, useSSL, cfg.MinioBucket)
		if err != nil {
			return nil, fmt.Errorf("minio public endpoint: %w", err)
		}
	}

	presignClient := mc
	if presignEndpoint != minioEndpoint || presignSecure != minioSecure {
		presignClient, err = minio.New(presignEndpoint, &minio.Options{
			Creds:        credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
			Secure:       presignSecure,
			Region:       cfg.MinioRegion,
			BucketLookup: bucketLookup,
		})
		if err != nil {
			return nil, fmt.Errorf("minio presign client: %w", err)
		}
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
		Client:        mc,
		PresignClient: presignClient,
		Bucket:        cfg.MinioBucket,
		uploadTTL:     uploadTTL,
		downloadTTL:   downloadTTL,
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

func normalizeEndpoint(raw string, defaultSecure bool, bucket string) (string, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false, fmt.Errorf("empty endpoint")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", false, err
		}
		if u.Host == "" {
			return "", false, fmt.Errorf("missing host")
		}
		path := strings.Trim(u.Path, "/")
		if path != "" {
			if bucket == "" {
				return "", false, fmt.Errorf("path is not allowed when bucket is empty")
			}
			if path != bucket {
				return "", false, fmt.Errorf("unexpected path in endpoint: %q", u.Path)
			}
		}
		if u.RawQuery != "" || u.Fragment != "" {
			return "", false, fmt.Errorf("query and fragment are not allowed in endpoint")
		}
		return u.Host, u.Scheme == "https", nil
	}

	return trimmed, defaultSecure, nil
}

func parseBucketLookup(raw string) (minio.BucketLookupType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto":
		return minio.BucketLookupAuto, nil
	case "dns", "virtual", "virtual-hosted":
		return minio.BucketLookupDNS, nil
	case "path", "path-style":
		return minio.BucketLookupPath, nil
	default:
		return minio.BucketLookupPath, fmt.Errorf("invalid MINIO_BUCKET_LOOKUP: %q", raw)
	}
}
