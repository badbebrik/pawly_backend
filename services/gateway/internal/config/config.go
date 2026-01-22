package config

import "os"

type Config struct {
	AppPort string

	JWTSecret string

	FileServiceGRPCAddr string
}

func Load() *Config {
	return &Config{
		AppPort:             getEnv("APP_PORT", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		FileServiceGRPCAddr: getEnv("FILE_SERVICE_GRPC_ADDR", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
