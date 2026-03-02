package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
)

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	JWTSecret           string
	JWTIssuer           string
	AccessTokenTTLMin   int
	RefreshTokenTTLDays int

	RabbitHost               string
	RabbitPort               string
	RabbitUser               string
	RabbitPassword           string
	RabbitNotificationsQueue string

	ProfileServiceURL       string
	InternalServiceToken    string
	GoogleOAuthClientID     string
	OAuthHTTPTimeoutSeconds int
}

func Load() *Config {
	cfg := &Config{
		AppPort:                  getEnv("APP_PORT", "8000"),
		PostgresUser:             getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:         getEnv("POSTGRES_PASSWORD", "password"),
		PostgresDB:               getEnv("POSTGRES_DB", "auth_db"),
		PostgresHost:             getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:             getEnv("POSTGRES_PORT", "5432"),
		RedisHost:                getEnv("REDIS_HOST", ""),
		RedisPort:                getEnv("REDIS_PORT", ""),
		RedisPassword:            getEnv("REDIS_PASSWORD", ""),
		RedisDB:                  getEnvInt("REDIS_DB", 0),
		JWTSecret:                getEnv("JWT_SECRET", "DEFAULT"),
		JWTIssuer:                getEnv("JWT_ISSUER", "pawly"),
		AccessTokenTTLMin:        getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15),
		RefreshTokenTTLDays:      getEnvInt("REFRESH_TOKEN_TTL_DAYS", 30),
		RabbitHost:               getEnv("RABBITMQ_HOST", "localhost"),
		RabbitPort:               getEnv("RABBITMQ_PORT", "5672"),
		RabbitUser:               getEnv("RABBITMQ_USER", ""),
		RabbitPassword:           getEnv("RABBITMQ_PASSWORD", ""),
		RabbitNotificationsQueue: getEnv("RABBITMQ_NOTIFICATIONS_QUEUE", ""),
		ProfileServiceURL:        getEnv("PROFILE_SERVICE_URL", "http://localhost:8001"),
		InternalServiceToken:     getEnv("INTERNAL_SERVICE_TOKEN", ""),
		GoogleOAuthClientID:      getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		OAuthHTTPTimeoutSeconds:  getEnvInt("OAUTH_HTTP_TIMEOUT_SECONDS", 5),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)

	if !ok || val == "" {
		return fallback
	}

	return val
}

func getEnvInt(key string, fallback int) int {
	valStr, ok := os.LookupEnv(key)

	if !ok || valStr == "" {
		return fallback
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatal().
			Str("env_key", key).
			Str("given_value", valStr).
			Err(err).
			Msg("Invalid integer value in environment variable")
	}

	return val
}
