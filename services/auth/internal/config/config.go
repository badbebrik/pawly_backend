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
	AccessTokenTTLMin   int
	RefreshTokenTTLDays int
}

func Load() *Config {
	cfg := &Config{
		AppPort:             getEnv("APP_PORT", "8000"),
		PostgresUser:        getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:    getEnv("POSTGRES_PASSWORD", "password"),
		PostgresDB:          getEnv("POSTGRES_DB", "auth_db"),
		PostgresHost:        getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:        getEnv("POSTGRES_PORT", "5432"),
		RedisHost:           getEnv("REDIS_ADDR", ""),
		RedisPort:           getEnv("REDIS_PORT", ""),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             getEnvInt("REDIS_DB", 0),
		JWTSecret:           getEnv("JWT_SECRET", "DEFAULT"),
		AccessTokenTTLMin:   getEnvInt("ACCESS_TOKEN_TTL_MINITUES", 15),
		RefreshTokenTTLDays: getEnvInt("REFRESH_TOKEN_TTL_DAYS", 30),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	log.Info().Str("val", val)

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
