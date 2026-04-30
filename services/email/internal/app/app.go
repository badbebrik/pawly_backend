package app

import (
	"context"
	"email/internal/config"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email/internal/queue"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config

	ctx    context.Context
	cancel context.CancelFunc

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

	consumer *queue.Consumer
	httpSrv  *http.Server
}

func New(cfg *config.Config) (*App, error) {
	rt, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		rabbitConn: rt.rabbitConn,
		rabbitCh:   rt.rabbitCh,
		consumer:   rt.consumer,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Email App resources...")

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
}

func (a *App) Run() error {
	if err := a.consumer.Start(a.ctx); err != nil {
		return err
	}
	log.Info().Str("queue", a.cfg.RabbitEmailQueue).Msg("started email notifications consumer")

	r := a.setupRoutes()
	a.httpSrv = &http.Server{
		Addr:    ":" + a.cfg.AppPort,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Email service...")
	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := a.httpSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cancel()
			return err
		}
		cancel()
	}
	return nil
}
