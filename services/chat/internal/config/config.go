package config

import "pawly/pkg/configenv"

type Config struct {
	AppPort          string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string
	ACLGRPCAddr      string
	ProfileGRPCAddr  string
	PetGRPCAddr      string
}

func Load() *Config {
	return &Config{
		AppPort:          configenv.String("APP_PORT", "8090"),
		PostgresUser:     configenv.String("POSTGRES_USER", ""),
		PostgresPassword: configenv.String("POSTGRES_PASSWORD", ""),
		PostgresDB:       configenv.String("POSTGRES_DB", ""),
		PostgresHost:     configenv.String("POSTGRES_HOST", "localhost"),
		PostgresPort:     configenv.String("POSTGRES_PORT", "5432"),
		ACLGRPCAddr:      configenv.String("ACL_GRPC_ADDR", ""),
		ProfileGRPCAddr:  configenv.String("PROFILE_GRPC_ADDR", ""),
		PetGRPCAddr:      configenv.String("PET_GRPC_ADDR", ""),
	}
}
