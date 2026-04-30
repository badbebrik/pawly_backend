package config

import "pawly/pkg/configenv"

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	RabbitHost      string
	RabbitPort      string
	RabbitUser      string
	RabbitPassword  string
	RabbitPushQueue string

	JWTSecret string
	JWTIssuer string

	FCMProjectID       string
	FCMCredentialsFile string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppPort: configenv.String("APP_PORT", "8090"),

		PostgresUser:     configenv.String("POSTGRES_USER", ""),
		PostgresPassword: configenv.String("POSTGRES_PASSWORD", ""),
		PostgresDB:       configenv.String("POSTGRES_DB", ""),
		PostgresHost:     configenv.String("POSTGRES_HOST", ""),
		PostgresPort:     configenv.String("POSTGRES_PORT", ""),

		RabbitHost:      configenv.String("RABBITMQ_HOST", ""),
		RabbitPort:      configenv.String("RABBITMQ_PORT", ""),
		RabbitUser:      configenv.String("RABBITMQ_USER", ""),
		RabbitPassword:  configenv.String("RABBITMQ_PASSWORD", ""),
		RabbitPushQueue: configenv.String("RABBITMQ_PUSH_QUEUE", "push.notifications"),

		JWTSecret: configenv.String("JWT_SECRET", "local_dev_jwt_secret"),
		JWTIssuer: configenv.String("JWT_ISSUER", "pawly"),

		FCMProjectID:       configenv.String("FCM_PROJECT_ID", ""),
		FCMCredentialsFile: configenv.String("FCM_CREDENTIALS_FILE", ""),
	}

	return cfg, nil
}
