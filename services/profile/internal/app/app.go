package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"profile/internal/config"
	"profile/internal/infrastructure/db"
	pgrepo "profile/internal/infrastructure/db/repository"
	"profile/internal/infrastructure/fileclient"
	"profile/internal/queue"
	"profile/internal/service"
)

type App struct {
	cfg *config.Config

	pg      *db.Postgres
	httpSrv *http.Server

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

	userEventsConsumer *queue.UserEventsConsumer
	fileClient         *fileclient.Client
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	fileClient, err := fileclient.New(cfg.FileServiceGRPCAddr)
	if err != nil {
		pg.Close()
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
		fileClient.Close()
		pg.Close()
		return nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fileClient.Close()
		pg.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if _, err := ch.QueueDeclare(
		cfg.RabbitUserEventsQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		fileClient.Close()
		pg.Close()
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitUserEventsQueue, err)
	}

	profileRepo := pgrepo.NewProfileRepository(pg.Pool)
	profileSvc := service.NewService(profileRepo, cfg, fileClient)

	userEventsConsumer := queue.NewUserEventsConsumer(ch, cfg.RabbitUserEventsQueue, profileSvc, cfg)

	app := &App{
		cfg:                cfg,
		pg:                 pg,
		fileClient:         fileClient,
		rabbitConn:         conn,
		rabbitCh:           ch,
		userEventsConsumer: userEventsConsumer,
	}

	r := app.setupRoutes(profileSvc)
	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	return app, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Profile App resources...")

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}

	if a.rabbitCh != nil {
		_ = a.rabbitCh.Close()
	}
	if a.rabbitConn != nil {
		_ = a.rabbitConn.Close()
	}
	if a.fileClient != nil {
		a.fileClient.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.userEventsConsumer.Start(ctx); err != nil {
		return err
	}
	log.Info().Str("queue", a.cfg.RabbitUserEventsQueue).Msg("started user events consumer")

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Profile service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}
