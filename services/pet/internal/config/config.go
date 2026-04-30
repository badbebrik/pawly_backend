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

	ACLGRPCAddr  string
	FileGRPCAddr string
}

func Load() (*Config, error) {
	return &Config{
		AppPort:          configenv.String("APP_PORT", "8085"),
		AppGRPCPort:      configenv.String("APP_GRPC_PORT", "50059"),
		PostgresUser:     configenv.String("POSTGRES_USER", "pet_user"),
		PostgresPassword: configenv.String("POSTGRES_PASSWORD", "supersecret"),
		PostgresDB:       configenv.String("POSTGRES_DB", "pet_db"),
		PostgresHost:     configenv.String("POSTGRES_HOST", "localhost"),
		PostgresPort:     configenv.String("POSTGRES_PORT", "5435"),
		ACLGRPCAddr:      configenv.String("ACL_GRPC_ADDR", "localhost:50057"),
		FileGRPCAddr:     configenv.String("FILE_GRPC_ADDR", "localhost:50056"),
	}, nil
}
