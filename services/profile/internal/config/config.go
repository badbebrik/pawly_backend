package config

import "pawly/pkg/configenv"

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

	FileServiceGRPCAddr  string
	InternalServiceToken string
}

func Load() (*Config, error) {
	return &Config{
		AppPort:              configenv.String("APP_PORT", "8086"),
		AppGRPCPort:          configenv.String("APP_GRPC_PORT", "50058"),
		PostgresUser:         configenv.String("POSTGRES_USER", ""),
		PostgresPassword:     configenv.String("POSTGRES_PASSWORD", ""),
		PostgresDB:           configenv.String("POSTGRES_DB", ""),
		PostgresHost:         configenv.String("POSTGRES_HOST", ""),
		PostgresPort:         configenv.String("POSTGRES_PORT", ""),
		DefaultLocale:        configenv.String("PROFILE_DEFAULT_LOCALE", "ru"),
		DefaultTimezone:      configenv.String("PROFILE_DEFAULT_TIMEZONE", "UTC"),
		FileServiceGRPCAddr:  configenv.String("FILE_SERVICE_GRPC_ADDR", ""),
		InternalServiceToken: configenv.String("INTERNAL_SERVICE_TOKEN", ""),
	}, nil
}
