package config

import "os"

type Config struct {
	AppHTTPPort string
	AppGRPCPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	InternalServiceToken string
}

func Load() *Config {
	return &Config{
		AppHTTPPort:          getEnv("APP_HTTP_PORT", "8087"),
		AppGRPCPort:          getEnv("APP_GRPC_PORT", "50057"),
		PostgresUser:         getEnv("POSTGRES_USER", "acl_user"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "supersecret"),
		PostgresDB:           getEnv("POSTGRES_DB", "acl_db"),
		PostgresHost:         getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:         getEnv("POSTGRES_PORT", "5434"),
		InternalServiceToken: getEnv("INTERNAL_SERVICE_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
