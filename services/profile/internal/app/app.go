package app

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"net/http"
	"profile/internal/config"
	"profile/internal/infrastructure/db"
	pgrepo "profile/internal/infrastructure/db/repository"
	"profile/internal/queue"
	"profile/internal/service"
	"time"
)

type App struct {
	cfg *config.Config

	pg      *db.Postgres
	httpSrv *http.Server

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

	userEventsConsumer *queue.UserEventsConsumer
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
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
		pg.Close()
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitUserEventsQueue, err)
	}

	profileRepo := pgrepo.NewProfileRepository(pg.Pool)
	profileSvc := service.NewService(profileRepo, cfg)

	userEventsConsumer := queue.NewUserEventsConsumer(ch, cfg.RabbitUserEventsQueue, profileSvc, cfg)

	app := &App{
		cfg:                cfg,
		pg:                 pg,
		rabbitConn:         conn,
		rabbitCh:           ch,
		userEventsConsumer: userEventsConsumer,
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
	if a.pg != nil {
		a.pg.Close()
	}
}
