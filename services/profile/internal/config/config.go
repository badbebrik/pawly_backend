package config

import "os"

type Config struct {
	AppPort     string
	AppGRPCPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	DefaultLocale   string
	DefaultTimezone string
	DefaultDateFmt  string

	FileServiceGRPCAddr string
}

func Load() *Config {
	return &Config{
		AppPort:             getEnv("APP_PORT", "8086"),
		AppGRPCPort:         getEnv("APP_GRPC_PORT", "50058"),
		PostgresUser:        getEnv("POSTGRES_USER", ""),
		PostgresPassword:    getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:          getEnv("POSTGRES_DB", ""),
		PostgresHost:        getEnv("POSTGRES_HOST", ""),
		PostgresPort:        getEnv("POSTGRES_PORT", ""),
		DefaultLocale:       getEnv("PROFILE_DEFAULT_LOCALE", ""),
		DefaultTimezone:     getEnv("PROFILE_DEFAULT_TIMEZONE", ""),
		DefaultDateFmt:      getEnv("PROFILE_DEFAULT_DATE_FORMAT", ""),
		FileServiceGRPCAddr: getEnv("FILE_SERVICE_GRPC_ADDR", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
