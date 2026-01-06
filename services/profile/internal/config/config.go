package config

import "os"

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	RabbitHost            string
	RabbitPort            string
	RabbitUser            string
	RabbitPassword        string
	RabbitUserEventsQueue string

	DefaultLocale   string
	DefaultTimezone string
	DefaultDateFmt  string
}

func Load() *Config {
	return &Config{
		AppPort:          getEnv("APP_PORT", ""),
		RabbitHost:       getEnv("RABBITMQ_HOST", ""),
		RabbitPort:       getEnv("RABBITMQ_PORT", ""),
		RabbitUser:       getEnv("RABBITMQ_USER", ""),
		RabbitPassword:   getEnv("RABBITMQ_PASSWORD", ""),
		PostgresUser:     getEnv("POSTGRES_USER", ""),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgresDB:       getEnv("POSTGRES_DB", ""),
		PostgresHost:     getEnv("POSTGRES_HOST", ""),
		PostgresPort:     getEnv("POSTGRES_PORT", ""),
		DefaultLocale:    getEnv("PROFILE_DEFAULT_LOCALE", ""),
		DefaultTimezone:  getEnv("PROFILE_DEFAULT_TIMEZONE", ""),
		DefaultDateFmt:   getEnv("PROFILE_DEFAULT_DATE_FORMAT", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
