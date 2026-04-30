package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pushuc "push/internal/application/usecase"
	"push/internal/config"
	"push/internal/infrastructure"
	"push/internal/queue"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg      *config.Config
	pg       *infrastructure.Postgres
	useCases *pushuc.Set

	ctx    context.Context
	cancel context.CancelFunc

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
	consumer   *queue.Consumer

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	rt, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	a := &App{
		cfg:        cfg,
		pg:         rt.pg,
		useCases:   rt.useCases,
		ctx:        ctx,
		cancel:     cancel,
		rabbitConn: rt.rabbitConn,
		rabbitCh:   rt.rabbitCh,
		consumer:   rt.consumer,
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
	log.Info().Str("queue", a.cfg.RabbitPushQueue).Msg("started push notifications consumer")

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
	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := a.httpSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cancel()
			return err
		}
		cancel()
	}
	a.cancel()
	return nil
}
