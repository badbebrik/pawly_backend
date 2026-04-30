package config

import "pawly/pkg/configenv"

type Config struct {
	AppPort string

	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	ACLGRPCAddr          string
	ACLHTTPBaseURL       string
	FileGRPCAddr         string
	JWTSecret            string
	JWTIssuer            string
	InternalServiceToken string

	RabbitHost                   string
	RabbitPort                   string
	RabbitUser                   string
	RabbitPassword               string
	RabbitPushQueue              string
	ScheduledDispatchIntervalSec int
	ScheduledDispatchBatchSize   int
	ScheduledHorizonIntervalSec  int
	ScheduledHorizonBatchSize    int
}

func Load() (*Config, error) {
	scheduledDispatchIntervalSec, err := configenv.Int("SCHEDULED_DISPATCH_INTERVAL_SEC", 60)
	if err != nil {
		return nil, err
	}
	scheduledDispatchBatchSize, err := configenv.Int("SCHEDULED_DISPATCH_BATCH_SIZE", 100)
	if err != nil {
		return nil, err
	}
	scheduledHorizonIntervalSec, err := configenv.Int("SCHEDULED_HORIZON_INTERVAL_SEC", 3600)
	if err != nil {
		return nil, err
	}
	scheduledHorizonBatchSize, err := configenv.Int("SCHEDULED_HORIZON_BATCH_SIZE", 100)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppPort:                      configenv.String("APP_PORT", "8088"),
		PostgresUser:                 configenv.String("POSTGRES_USER", "health_user"),
		PostgresPassword:             configenv.String("POSTGRES_PASSWORD", "supersecret"),
		PostgresDB:                   configenv.String("POSTGRES_DB", "health_db"),
		PostgresHost:                 configenv.String("POSTGRES_HOST", "localhost"),
		PostgresPort:                 configenv.String("POSTGRES_PORT", "5437"),
		ACLGRPCAddr:                  configenv.String("ACL_GRPC_ADDR", "localhost:50057"),
		ACLHTTPBaseURL:               configenv.String("ACL_HTTP_BASE_URL", "http://localhost:8087"),
		FileGRPCAddr:                 configenv.String("FILE_GRPC_ADDR", "localhost:50056"),
		JWTSecret:                    configenv.String("JWT_SECRET", "local_dev_jwt_secret"),
		JWTIssuer:                    configenv.String("JWT_ISSUER", "pawly"),
		InternalServiceToken:         configenv.String("INTERNAL_SERVICE_TOKEN", ""),
		RabbitHost:                   configenv.String("RABBITMQ_HOST", "localhost"),
		RabbitPort:                   configenv.String("RABBITMQ_PORT", "5672"),
		RabbitUser:                   configenv.String("RABBITMQ_USER", "auth_user"),
		RabbitPassword:               configenv.String("RABBITMQ_PASSWORD", "supersecret"),
		RabbitPushQueue:              configenv.String("RABBITMQ_PUSH_QUEUE", "push.notifications"),
		ScheduledDispatchIntervalSec: scheduledDispatchIntervalSec,
		ScheduledDispatchBatchSize:   scheduledDispatchBatchSize,
		ScheduledHorizonIntervalSec:  scheduledHorizonIntervalSec,
		ScheduledHorizonBatchSize:    scheduledHorizonBatchSize,
	}

	return cfg, nil
}
