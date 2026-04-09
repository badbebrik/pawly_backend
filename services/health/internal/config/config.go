package config

import "os"

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	ACLGRPCAddr  string
	ACLHTTPBaseURL string
	FileGRPCAddr string
	JWTSecret    string
	JWTIssuer    string
	InternalServiceToken string

	RabbitHost          string
	RabbitPort          string
	RabbitUser          string
	RabbitPassword      string
	RabbitPushJobsQueue string
	ScheduledDispatchIntervalSec string
	ScheduledDispatchBatchSize string
}

func Load() *Config {
	return &Config{
		AppPort:          getEnv("APP_PORT", "8088"),
		PostgresUser:     getEnv("POSTGRES_USER", "health_user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "supersecret"),
		PostgresDB:       getEnv("POSTGRES_DB", "health_db"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5437"),
		ACLGRPCAddr:      getEnv("ACL_GRPC_ADDR", "localhost:50057"),
		ACLHTTPBaseURL:   getEnv("ACL_HTTP_BASE_URL", "http://localhost:8087"),
		FileGRPCAddr:     getEnv("FILE_GRPC_ADDR", "localhost:50056"),
		JWTSecret:        getEnv("JWT_SECRET", "local_dev_jwt_secret"),
		JWTIssuer:        getEnv("JWT_ISSUER", "pawly"),
		InternalServiceToken: getEnv("INTERNAL_SERVICE_TOKEN", ""),
		RabbitHost:       getEnv("RABBITMQ_HOST", "localhost"),
		RabbitPort:       getEnv("RABBITMQ_PORT", "5672"),
		RabbitUser:       getEnv("RABBITMQ_USER", "auth_user"),
		RabbitPassword:   getEnv("RABBITMQ_PASSWORD", "supersecret"),
		RabbitPushJobsQueue: getEnv("RABBITMQ_PUSH_JOBS_QUEUE", "push.jobs"),
		ScheduledDispatchIntervalSec: getEnv("SCHEDULED_DISPATCH_INTERVAL_SEC", "60"),
		ScheduledDispatchBatchSize: getEnv("SCHEDULED_DISPATCH_BATCH_SIZE", "100"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
