package config

import (
	"pawly/pkg/configenv"
)

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	MinioEndpoint     string
	MinioAccessKey    string
	MinioSecretKey    string
	MinioBucket       string
	MinioUseSSL       bool
	MinioRegion       string
	MinioBucketLookup string

	MinioPublicEndpoint   string
	MinioSkipBucketEnsure bool

	UploadURLTTLSeconds    int
	DownloadURLTTLSeconds  int
	CleanupIntervalSeconds int
	CleanupBatchSize       int
}

func Load() (*Config, error) {
	minioUseSSL, err := configenv.Bool("MINIO_USE_SSL", false)
	if err != nil {
		return nil, err
	}
	minioSkipBucketEnsure, err := configenv.Bool("MINIO_SKIP_BUCKET_ENSURE", false)
	if err != nil {
		return nil, err
	}
	uploadURLTTLSeconds, err := configenv.Int("UPLOAD_URL_TTL_SECONDS", 900)
	if err != nil {
		return nil, err
	}
	downloadURLTTLSeconds, err := configenv.Int("DOWNLOAD_URL_TTL_SECONDS", 900)
	if err != nil {
		return nil, err
	}
	cleanupIntervalSeconds, err := configenv.Int("CLEANUP_INTERVAL_SECONDS", 300)
	if err != nil {
		return nil, err
	}
	cleanupBatchSize, err := configenv.Int("CLEANUP_BATCH_SIZE", 100)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppPort:                configenv.String("APP_PORT", ""),
		PostgresUser:           configenv.String("POSTGRES_USER", ""),
		PostgresPassword:       configenv.String("POSTGRES_PASSWORD", ""),
		PostgresDB:             configenv.String("POSTGRES_DB", ""),
		PostgresHost:           configenv.String("POSTGRES_HOST", ""),
		PostgresPort:           configenv.String("POSTGRES_PORT", ""),
		MinioEndpoint:          configenv.String("MINIO_ENDPOINT", ""),
		MinioAccessKey:         configenv.String("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:         configenv.String("MINIO_SECRET_KEY", ""),
		MinioBucket:            configenv.String("MINIO_BUCKET", ""),
		MinioUseSSL:            minioUseSSL,
		MinioRegion:            configenv.String("MINIO_REGION", "us-east-1"),
		MinioBucketLookup:      configenv.String("MINIO_BUCKET_LOOKUP", "path"),
		MinioPublicEndpoint:    configenv.String("MINIO_PUBLIC_ENDPOINT", ""),
		MinioSkipBucketEnsure:  minioSkipBucketEnsure,
		UploadURLTTLSeconds:    uploadURLTTLSeconds,
		DownloadURLTTLSeconds:  downloadURLTTLSeconds,
		CleanupIntervalSeconds: cleanupIntervalSeconds,
		CleanupBatchSize:       cleanupBatchSize,
	}

	return cfg, nil
}
