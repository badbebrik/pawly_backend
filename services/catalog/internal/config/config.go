package config

import (
	"os"
)

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	AdminToken string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8090"),

		PostgresUser:     getEnv("POSTGRES_USER", ""),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:       getEnv("POSTGRES_DB", ""),
		PostgresHost:     getEnv("POSTGRES_HOST", ""),
		PostgresPort:     getEnv("POSTGRES_PORT", ""),
		AdminToken:       getEnv("ADMIN_TOKEN", ""),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
