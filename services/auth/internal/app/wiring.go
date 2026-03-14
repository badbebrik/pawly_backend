package app

import (
	"auth/internal/application/usecase"
	"auth/internal/config"
	"auth/internal/infrastructure/db"
	pgrepo "auth/internal/infrastructure/db/repository"
	"auth/internal/infrastructure/oauth"
	"auth/internal/infrastructure/outbox"
	"auth/internal/infrastructure/profileclient"
	"auth/internal/infrastructure/rabbit"
	"auth/internal/infrastructure/redis"
	"auth/internal/infrastructure/redis/redisstore"
	"auth/internal/infrastructure/tokens"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type runtime struct {
	authSvc    *usecase.Set
	pg         *db.Postgres
	redis      *redisdb.Redis
	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
	profile    *profileclient.GRPCClient
	outboxWkr  *outbox.Worker
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	redis, err := redisdb.NewRedis(cfg)
	if err != nil {
		pg.Close()
		return nil, err
	}

	rabbitConn, rabbitCh, err := newRabbitChannel(cfg)
	if err != nil {
		_ = redis.Close()
		pg.Close()
		return nil, err
	}

	profile, err := profileclient.New(cfg.ProfileServiceGRPCAddr)
	if err != nil {
		_ = rabbitCh.Close()
		_ = rabbitConn.Close()
		_ = redis.Close()
		pg.Close()
		return nil, err
	}

	authSvc, outboxWkr := buildAuthModule(cfg, pg, redis, rabbitCh, profile)

	return &runtime{
		authSvc:    authSvc,
		pg:         pg,
		redis:      redis,
		rabbitConn: rabbitConn,
		rabbitCh:   rabbitCh,
		profile:    profile,
		outboxWkr:  outboxWkr,
	}, nil
}

func buildAuthModule(cfg *config.Config, pg *db.Postgres, redis *redisdb.Redis, rabbitCh *amqp091.Channel, profile *profileclient.GRPCClient) (*usecase.Set, *outbox.Worker) {
	userRepo := pgrepo.NewUserRepo(pg.Pool)
	sessionRepo := pgrepo.NewSessionRepo(pg.Pool)
	oauthRepo := pgrepo.NewOAuthIdentityRepo(pg.Pool)
	outboxRepo := pgrepo.NewOutboxRepo(pg.Pool)
	resetTokenRepo := redisdb.NewResetTokenStore(redis.Client())
	verificationRepo := redisstore.NewRedisStore(redis.Client())

	outboxPublisher := outbox.NewPublisher(outboxRepo)
	rabbitPublisher := rabbit.NewRabbitPublisher(rabbitCh, cfg.RabbitNotificationsQueue)
	outboxWorker := outbox.NewWorker(
		outboxRepo,
		rabbitPublisher,
		time.Duration(cfg.OutboxWorkerIntervalMS)*time.Millisecond,
		cfg.OutboxWorkerBatchSize,
	)

	authSvc := usecase.NewSet(usecase.Dependencies{
		Users:        userRepo,
		Sessions:     sessionRepo,
		OAuth:        oauthRepo,
		ResetTokens:  resetTokenRepo,
		Verification: verificationRepo,
		Notifier:     outboxPublisher,
		Tokens:       tokens.NewJWTService(*cfg),
		Profiles:     profile,
		OAuthVerify:  oauth.NewGoogleVerifier(time.Duration(cfg.OAuthHTTPTimeoutSeconds)*time.Second, cfg.GoogleOAuthClientID),
	})

	return authSvc, outboxWorker
}

func newRabbitChannel(cfg *config.Config) (*amqp091.Connection, *amqp091.Channel, error) {
	conn, err := amqp091.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitHost,
		cfg.RabbitPort,
	))
	if err != nil {
		return nil, nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if _, err := ch.QueueDeclare(
		cfg.RabbitNotificationsQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, fmt.Errorf("rabbit queue declare: %w", err)
	}

	return conn, ch, nil
}
