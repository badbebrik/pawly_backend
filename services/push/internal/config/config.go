package config

import "os"

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	RabbitHost          string
	RabbitPort          string
	RabbitUser          string
	RabbitPassword      string
	RabbitPushJobsQueue string

	JWTSecret string
	JWTIssuer string

	FCMProjectID       string
	FCMCredentialsFile string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8090"),

		PostgresUser:     getEnv("POSTGRES_USER", ""),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:       getEnv("POSTGRES_DB", ""),
		PostgresHost:     getEnv("POSTGRES_HOST", ""),
		PostgresPort:     getEnv("POSTGRES_PORT", ""),

		RabbitHost:          getEnv("RABBITMQ_HOST", ""),
		RabbitPort:          getEnv("RABBITMQ_PORT", ""),
		RabbitUser:          getEnv("RABBITMQ_USER", ""),
		RabbitPassword:      getEnv("RABBITMQ_PASSWORD", ""),
		RabbitPushJobsQueue: getEnv("RABBITMQ_PUSH_JOBS_QUEUE", "push.jobs"),

		JWTSecret: getEnv("JWT_SECRET", "local_dev_jwt_secret"),
		JWTIssuer: getEnv("JWT_ISSUER", "pawly"),

		FCMProjectID:       getEnv("FCM_PROJECT_ID", ""),
		FCMCredentialsFile: getEnv("FCM_CREDENTIALS_FILE", ""),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
