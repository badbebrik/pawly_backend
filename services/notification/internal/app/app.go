package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"net/http"
	"notification/internal/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	config *config.Config

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

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

	return &App{
		config:     cfg,
		rabbitConn: conn,
		rabbitCh:   ch,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Notification App resources...")

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
