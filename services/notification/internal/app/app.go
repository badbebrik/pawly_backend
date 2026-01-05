package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"net/http"
	"notification/internal/config"
	"notification/internal/handler"
	"notification/internal/queue"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	config *config.Config

	authCtx    context.Context
	authCancel context.CancelFunc

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

	consumer *queue.Consumer

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
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
		_ = conn.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if _, err := ch.QueueDeclare(cfg.RabbitEventsQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("queue declare events: %w", err)
	}
	if _, err := ch.QueueDeclare(cfg.RabbitEmailJobsQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("queue declare email jobs: %w", err)
	}
	if _, err := ch.QueueDeclare(cfg.RabbitPushJobsQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("queue declare push jobs: %w", err)
	}

	emailPublisher := queue.NewEmailPublisher(ch, cfg.RabbitEmailJobsQueue)
	evHandler := handler.NewEventHandler(emailPublisher)
	consumer := queue.NewConsumer(ch, cfg.RabbitEventsQueue, evHandler)

	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		config:     cfg,
		authCtx:    ctx,
		authCancel: cancel,
		rabbitConn: conn,
		rabbitCh:   ch,
		consumer:   consumer,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Notification App resources...")

	if a.authCancel != nil {
		a.authCancel()
	}

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}

	if a.rabbitCh != nil {
		if err := a.rabbitCh.Close(); err != nil {
			log.Error().Err(err).Msg("rabbit channel close failed")
		}
	}
	if a.rabbitConn != nil {
		if err := a.rabbitConn.Close(); err != nil {
			log.Error().Err(err).Msg("rabbit conn close failed")
		}
	}
}

func (a *App) Run() error {
	if err := a.consumer.Start(a.authCtx); err != nil {
		return err
	}
	log.Info().Str("queue", a.config.RabbitEventsQueue).Msg("started rabbit consumer")

	r := a.setupRoutes()
	a.httpSrv = &http.Server{
		Addr:    ":" + a.config.AppPort,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", a.config.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Notification service...")

	return nil
}
