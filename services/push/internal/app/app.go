package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"push/internal/config"
	"push/internal/handler"
	"push/internal/infrastructure"
	pgrepo "push/internal/infrastructure/repository"
	"push/internal/queue"
	"push/internal/sender"
	"push/internal/service"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config
	pg  *infrastructure.Postgres
	svc *service.Service

	ctx    context.Context
	cancel context.CancelFunc

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
	consumer   *queue.Consumer
	sender     sender.Sender

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	pg, err := infrastructure.NewPostgres(cfg)
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
		pg.Close()
		return nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		pg.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}
	if _, err := ch.QueueDeclare(cfg.RabbitPushJobsQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		pg.Close()
		return nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitPushJobsQueue, err)
	}

	repo := pgrepo.NewPushRepository(pg.Pool)
	ctx, cancel := context.WithCancel(context.Background())
	svc := service.New(repo)
	pushSender := buildSender(cfg)
	jobHandler := handler.NewPushJobHandler(svc, pushSender)
	a := &App{
		cfg:        cfg,
		pg:         pg,
		svc:        svc,
		ctx:        ctx,
		cancel:     cancel,
		rabbitConn: conn,
		rabbitCh:   ch,
		consumer:   queue.NewConsumer(ch, cfg.RabbitPushJobsQueue, jobHandler),
		sender:     pushSender,
	}
	a.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: a.setupRoutes(),
	}
	return a, nil
}

func (a *App) Close() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
	}
	if a.rabbitCh != nil {
		_ = a.rabbitCh.Close()
	}
	if a.rabbitConn != nil {
		_ = a.rabbitConn.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	if err := a.consumer.Start(a.ctx); err != nil {
		return err
	}
	log.Info().Str("queue", a.cfg.RabbitPushJobsQueue).Msg("started push.jobs consumer")

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting Push HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("push http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Push service...")
	return nil
}

func buildSender(cfg *config.Config) sender.Sender {
	if cfg.FCMProjectID == "" || cfg.FCMCredentialsFile == "" {
		log.Warn().Msg("FCM sender is not configured, using noop sender")
		return sender.NewNoopSender()
	}

	fcmSender, err := sender.NewFCMSender(cfg.FCMProjectID, cfg.FCMCredentialsFile)
	if err != nil {
		log.Error().Err(err).Msg("failed to init FCM sender, using noop sender")
		return sender.NewNoopSender()
	}
	return fcmSender
}
