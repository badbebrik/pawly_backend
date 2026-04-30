package app

import (
	"auth/internal/application/usecase"
	"auth/internal/config"
	"auth/internal/infrastructure/db"
	"auth/internal/infrastructure/outbox"
	"auth/internal/infrastructure/profileclient"
	"auth/internal/infrastructure/redis"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg      *config.Config
	useCases *usecase.Set
	httpSrv  *http.Server

	pg         *db.Postgres
	redis      *redisdb.Redis
	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
	profile    *profileclient.GRPCClient
	outboxWkr  *outbox.Worker
}

func New(cfg *config.Config) (*App, error) {
	runtime, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	a := &App{
		cfg:        cfg,
		useCases:   runtime.useCases,
		pg:         runtime.pg,
		redis:      runtime.redis,
		rabbitConn: runtime.rabbitConn,
		rabbitCh:   runtime.rabbitCh,
		profile:    runtime.profile,
		outboxWkr:  runtime.outboxWkr,
	}

	a.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: a.setupRoutes(),
	}

	return a, nil
}

func (a *App) Close() {
	log.Info().Msg("closing App resources...")

	if a.redis != nil {
		_ = a.redis.Close()
	}

	if a.pg != nil {
		a.pg.Close()
	}

	if a.rabbitCh != nil {
		_ = a.rabbitCh.Close()
	}
	if a.rabbitConn != nil {
		_ = a.rabbitConn.Close()
	}
	if a.profile != nil {
		a.profile.Close()
	}
}

func (a *App) Run() error {
	go func() {
		log.Info().
			Str("port", a.cfg.AppPort).
			Msg("starting HTTP server")

		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server crash")
		}
	}()

	runCtx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()

	if a.outboxWkr != nil {
		go a.outboxWkr.Run(runCtx)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down HTTP server...")
	cancelRun()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
