package app

import (
	"fmt"

	pushuc "push/internal/application/usecase"
	"push/internal/config"
	"push/internal/handler"
	"push/internal/infrastructure"
	pgrepo "push/internal/infrastructure/repository"
	"push/internal/queue"
	"push/internal/sender"

	"github.com/rabbitmq/amqp091-go"
)

type runtime struct {
	pg         *infrastructure.Postgres
	useCases   *pushuc.Set
	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
	consumer   *queue.Consumer
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	pg, err := infrastructure.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	rabbitConn, rabbitCh, err := newRabbitChannel(cfg)
	if err != nil {
		pg.Close()
		return nil, err
	}

	repo := pgrepo.NewPushRepository(pg.Pool)
	useCases := pushuc.New(repo)
	pushSender := buildSender(cfg)
	jobHandler := handler.NewPushJobHandler(useCases, pushSender)

	return &runtime{
		pg:         pg,
		useCases:   useCases,
		rabbitConn: rabbitConn,
		rabbitCh:   rabbitCh,
		consumer:   queue.NewConsumer(rabbitCh, cfg.RabbitPushQueue, jobHandler),
	}, nil
}

func newRabbitChannel(cfg *config.Config) (*amqp091.Connection, *amqp091.Channel, error) {
	conn, err := amqp091.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitHost,
		cfg.RabbitPort,
	))
	if err != nil {
		return nil, nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if _, err := ch.QueueDeclare(cfg.RabbitPushQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitPushQueue, err)
	}

	return conn, ch, nil
}

func buildSender(cfg *config.Config) sender.Sender {
	if cfg.FCMProjectID == "" || cfg.FCMCredentialsFile == "" {
		return sender.NewNoopSender()
	}

	fcmSender, err := sender.NewFCMSender(cfg.FCMProjectID, cfg.FCMCredentialsFile)
	if err != nil {
		return sender.NewNoopSender()
	}
	return fcmSender
}
