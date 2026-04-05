package config

import "os"

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
	MinioUseSSL       string
	MinioRegion       string
	MinioBucketLookup string

	MinioPublicEndpoint   string
	MinioSkipBucketEnsure string

	UploadURLTTLSeconds   string
	DownloadURLTTLSeconds string
	CleanupIntervalSeconds string
	CleanupBatchSize       string
}

func Load() *Config {
	return &Config{
		AppPort:               getEnv("APP_PORT", ""),
		PostgresUser:          getEnv("POSTGRES_USER", ""),
		PostgresPassword:      getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:            getEnv("POSTGRES_DB", ""),
		PostgresHost:          getEnv("POSTGRES_HOST", ""),
		PostgresPort:          getEnv("POSTGRES_PORT", ""),
		MinioEndpoint:         getEnv("MINIO_ENDPOINT", ""),
		MinioAccessKey:        getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:        getEnv("MINIO_SECRET_KEY", ""),
		MinioBucket:           getEnv("MINIO_BUCKET", ""),
		MinioUseSSL:           getEnv("MINIO_USE_SSL", ""),
		MinioRegion:           getEnv("MINIO_REGION", "us-east-1"),
		MinioBucketLookup:     getEnv("MINIO_BUCKET_LOOKUP", "path"),
		MinioPublicEndpoint:   getEnv("MINIO_PUBLIC_ENDPOINT", ""),
		MinioSkipBucketEnsure: getEnv("MINIO_SKIP_BUCKET_ENSURE", "false"),
		UploadURLTTLSeconds:   getEnv("UPLOAD_URL_TTL_SECONDS", ""),
		DownloadURLTTLSeconds: getEnv("DOWNLOAD_URL_TTL_SECONDS", ""),
		CleanupIntervalSeconds: getEnv("CLEANUP_INTERVAL_SECONDS", "300"),
		CleanupBatchSize:       getEnv("CLEANUP_BATCH_SIZE", "100"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
