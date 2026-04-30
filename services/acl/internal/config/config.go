package config

import (
	"pawly/pkg/configenv"
)

type Config struct {
	AppHTTPPort string
	AppGRPCPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	InternalServiceToken   string
	ProfileServiceGRPCAddr string
	PetServiceGRPCAddr     string
	InviteTTLMinutes       int
	InviteDeeplinkBase     string
}

func Load() (*Config, error) {
	inviteTTLMinutes, err := configenv.Int("INVITE_TTL_MINUTES", 10080)
	if err != nil {
		return nil, err
	}
	return &Config{
		AppHTTPPort:            configenv.String("APP_HTTP_PORT", "8087"),
		AppGRPCPort:            configenv.String("APP_GRPC_PORT", "50057"),
		PostgresUser:           configenv.String("POSTGRES_USER", "acl_user"),
		PostgresPassword:       configenv.String("POSTGRES_PASSWORD", "supersecret"),
		PostgresDB:             configenv.String("POSTGRES_DB", "acl_db"),
		PostgresHost:           configenv.String("POSTGRES_HOST", "localhost"),
		PostgresPort:           configenv.String("POSTGRES_PORT", "5434"),
		InternalServiceToken:   configenv.String("INTERNAL_SERVICE_TOKEN", ""),
		ProfileServiceGRPCAddr: configenv.String("PROFILE_SERVICE_GRPC_ADDR", "localhost:50058"),
		PetServiceGRPCAddr:     configenv.String("PET_SERVICE_GRPC_ADDR", "localhost:50059"),
		InviteTTLMinutes:       inviteTTLMinutes,
		InviteDeeplinkBase:     configenv.String("INVITE_DEEPLINK_BASE", "pawly://invite?token="),
	}, nil
}
