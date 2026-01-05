package config

import "os"

type Config struct {
	AppPort string

	RabbitHost     string
	RabbitPort     string
	RabbitUser     string
	RabbitPassword string

	EventsQueue    string
	EmailJobsQueue string
}

func Load() *Config {
	return &Config{
		AppPort: os.Getenv("APP_PORT"),

		RabbitHost:     os.Getenv("RABBITMQ_HOST"),
		RabbitPort:     os.Getenv("RABBITMQ_PORT"),
		RabbitUser:     os.Getenv("RABBITMQ_USER"),
		RabbitPassword: os.Getenv("RABBITMQ_PASSWORD"),

		EventsQueue:    os.Getenv("RABBITMQ_EVENTS_QUEUE"),
		EmailJobsQueue: os.Getenv("RABBITMQ_EMAIL_JOBS_QUEUE"),
	}
}
