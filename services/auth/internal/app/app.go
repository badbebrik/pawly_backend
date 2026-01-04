package app

import (
	"auth/internal/config"
	"auth/internal/infrastructure/db"
	pgrepo "auth/internal/infrastructure/db/repository"
	"auth/internal/infrastructure/rabbit"
	"auth/internal/infrastructure/redis"
	"auth/internal/infrastructure/redis/redisstore"
	"auth/internal/infrastructure/tokens"
	authsvc "auth/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	Config                *config.Config
	PG                    *db.Postgres
	Redis                 *redisdb.Redis
	NotificationPublisher rabbit.Publisher
	AuthSvc               *authsvc.Service
	RabbitConn            *amqp091.Connection
	RabbitCh              *amqp091.Channel
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	redis, err := redisdb.NewRedis(cfg)
	if err != nil {
		return nil, err
	}

	conn, err := amqp091.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitHost,
		cfg.RabbitPort,
	))
	if err != nil {
		return nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		cfg.RabbitNotificationsQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbit queue declare: %w", err)
	}

	publisher := rabbit.NewRabbitPublisher(ch, cfg.RabbitNotificationsQueue)

	userRepo := pgrepo.NewUserRepo(pg.Pool)
	sessionRepo := pgrepo.NewSessionRepo(pg.Pool)
	oauthRepo := pgrepo.NewOAuthIdentityRepo(pg.Pool)
	deviceRepo := pgrepo.NewUserDeviceRepo(pg.Pool)
	verificationRepo := redisstore.NewRedisStore(redis.Client())

	jwtSvc := tokens.NewJWTService(*cfg)

	authSvc := authsvc.NewService(
		userRepo,
		sessionRepo,
		oauthRepo,
		deviceRepo,
		verificationRepo,
		publisher,
		jwtSvc,
	)

	return &App{
		Config:                cfg,
		PG:                    pg,
		Redis:                 redis,
		NotificationPublisher: publisher,
		AuthSvc:               authSvc,
		RabbitConn:            conn,
		RabbitCh:              ch,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing App resources...")

	if a.Redis != nil {
		err := a.Redis.Close()
		if err != nil {
			return
		}
	}

	if a.PG != nil {
		a.PG.Close()
	}

	if a.RabbitCh != nil {
		_ = a.RabbitCh.Close()
	}
	if a.RabbitConn != nil {
		_ = a.RabbitConn.Close()
	}
}

func (a *App) Run() error {
	r := a.setupRoutes()

	srv := &http.Server{
		Addr:    ":" + a.Config.AppPort,
		Handler: r,
	}

	go func() {
		log.Info().
			Str("port", a.Config.AppPort).
			Msg("starting HTTP server")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
