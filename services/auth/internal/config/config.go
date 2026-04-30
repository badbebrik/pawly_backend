package config

import "pawly/pkg/configenv"

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

	JWTSecret                string
	JWTIssuer                string
	AccessTokenTTLMin        int
	RefreshTokenTTLDays      int
	PasswordResetTokenTTLMin int
	OutboxWorkerIntervalMS   int
	OutboxWorkerBatchSize    int

	RabbitHost       string
	RabbitPort       string
	RabbitUser       string
	RabbitPassword   string
	RabbitEmailQueue string

	ProfileServiceGRPCAddr  string
	GoogleOAuthClientID     string
	OAuthHTTPTimeoutSeconds int
}

func Load() (*Config, error) {
	appPort := configenv.String("APP_PORT", "8000")
	postgresUser := configenv.String("POSTGRES_USER", "postgres")
	postgresPassword := configenv.String("POSTGRES_PASSWORD", "password")
	postgresDB := configenv.String("POSTGRES_DB", "auth_db")
	postgresHost := configenv.String("POSTGRES_HOST", "localhost")
	postgresPort := configenv.String("POSTGRES_PORT", "5432")
	redisHost := configenv.String("REDIS_HOST", "")
	redisPort := configenv.String("REDIS_PORT", "")
	redisPassword := configenv.String("REDIS_PASSWORD", "")
	redisDB, err := configenv.Int("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}
	jwtSecret := configenv.String("JWT_SECRET", "DEFAULT")
	jwtIssuer := configenv.String("JWT_ISSUER", "pawly")
	accessTokenTTLMin, err := configenv.Int("ACCESS_TOKEN_TTL_MINUTES", 15)
	if err != nil {
		return nil, err
	}
	refreshTokenTTLDays, err := configenv.Int("REFRESH_TOKEN_TTL_DAYS", 30)
	if err != nil {
		return nil, err
	}
	passwordResetTokenTTLMin, err := configenv.Int("PASSWORD_RESET_TOKEN_TTL_MINUTES", 15)
	if err != nil {
		return nil, err
	}
	outboxWorkerIntervalMS, err := configenv.Int("OUTBOX_WORKER_INTERVAL_MS", 2000)
	if err != nil {
		return nil, err
	}
	outboxWorkerBatchSize, err := configenv.Int("OUTBOX_WORKER_BATCH_SIZE", 100)
	if err != nil {
		return nil, err
	}
	rabbitHost := configenv.String("RABBITMQ_HOST", "localhost")
	rabbitPort := configenv.String("RABBITMQ_PORT", "5672")
	rabbitUser := configenv.String("RABBITMQ_USER", "")
	rabbitPassword := configenv.String("RABBITMQ_PASSWORD", "")
	rabbitEmailQueue := configenv.String("RABBITMQ_EMAIL_QUEUE", "email.notifications")
	profileServiceGRPCAddr := configenv.String("PROFILE_SERVICE_GRPC_ADDR", "localhost:50058")
	googleOAuthClientID, err := configenv.RequiredString("GOOGLE_OAUTH_CLIENT_ID")
	if err != nil {
		return nil, err
	}
	oauthHTTPTimeoutSeconds, err := configenv.Int("OAUTH_HTTP_TIMEOUT_SECONDS", 5)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppPort:                  appPort,
		PostgresUser:             postgresUser,
		PostgresPassword:         postgresPassword,
		PostgresDB:               postgresDB,
		PostgresHost:             postgresHost,
		PostgresPort:             postgresPort,
		RedisHost:                redisHost,
		RedisPort:                redisPort,
		RedisPassword:            redisPassword,
		RedisDB:                  redisDB,
		JWTSecret:                jwtSecret,
		JWTIssuer:                jwtIssuer,
		AccessTokenTTLMin:        accessTokenTTLMin,
		RefreshTokenTTLDays:      refreshTokenTTLDays,
		PasswordResetTokenTTLMin: passwordResetTokenTTLMin,
		OutboxWorkerIntervalMS:   outboxWorkerIntervalMS,
		OutboxWorkerBatchSize:    outboxWorkerBatchSize,
		RabbitHost:               rabbitHost,
		RabbitPort:               rabbitPort,
		RabbitUser:               rabbitUser,
		RabbitPassword:           rabbitPassword,
		RabbitEmailQueue:         rabbitEmailQueue,
		ProfileServiceGRPCAddr:   profileServiceGRPCAddr,
		GoogleOAuthClientID:      googleOAuthClientID,
		OAuthHTTPTimeoutSeconds:  oauthHTTPTimeoutSeconds,
	}

	return cfg, nil
}
