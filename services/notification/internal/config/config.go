package config

import "os"

type Config struct {
	AppPort string

	RabbitHost     string
	RabbitPort     string
	RabbitUser     string
	RabbitPassword string

	RabbitEventsQueue    string
	RabbitEmailJobsQueue string
	RabbitPushJobsQueue  string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8081"),

		RabbitHost:     getEnv("RABBITMQ_HOST", ""),
		RabbitPort:     getEnv("RABBITMQ_PORT", ""),
		RabbitUser:     getEnv("RABBITMQ_USER", ""),
		RabbitPassword: getEnv("RABBITMQ_PASSWORD", ""),

		RabbitEventsQueue:    getEnv("RABBITMQ_EVENTS_QUEUE", ""),
		RabbitEmailJobsQueue: getEnv("RABBITMQ_EMAIL_JOBS_QUEUE", ""),
		RabbitPushJobsQueue:  getEnv("RABBITMQ_PUSH_JOBS_QUEUE", ""),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
