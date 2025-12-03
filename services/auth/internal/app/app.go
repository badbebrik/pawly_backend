package app

import (
	"auth/internal/config"
	"auth/internal/db"
	"auth/internal/notifications"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type App struct {
	Config                *config.Config
	PG                    *db.Postgres
	Redis                 *db.Redis
	NotificationPublisher notifications.Publisher
}

func New(cfg *config.Config) (*App, error) {
	pg, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	redis, err := db.NewRedis(cfg)
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

	publisher := notifications.NewRabbitPublisher(ch, cfg.RabbitNotificationsQueue)

	return &App{
		Config:                cfg,
		PG:                    pg,
		Redis:                 redis,
		NotificationPublisher: publisher,
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
}
