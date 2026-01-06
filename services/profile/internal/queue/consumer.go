package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"profile/internal/config"
	"profile/internal/service"
)

type UserEventsConsumer struct {
	ch        *amqp091.Channel
	queueName string
	svc       *service.ProfileService
	cfg       *config.Config
}

func NewUserEventsConsumer(ch *amqp091.Channel, queueName string, svc *service.ProfileService, cfg *config.Config) *UserEventsConsumer {
	return &UserEventsConsumer{
		ch:        ch,
		queueName: queueName,
		svc:       svc,
		cfg:       cfg,
	}
}

func (c *UserEventsConsumer) Start(ctx context.Context) error {
	if err := c.ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	deliveries, err := c.ch.Consume(
		c.queueName,
		"profile-user-events-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume %s: %w", c.queueName, err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-deliveries:
				if !ok {
					return
				}
				c.handleMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *UserEventsConsumer) handleMessage(ctx context.Context, msg amqp091.Delivery) {
	var ev UserCreatedEvent
	if err := json.Unmarshal(msg.Body, &ev); err != nil {
		log.Warn().Err(err).Msg("invalid user event json")
		_ = msg.Ack(false)
		return
	}

	if ev.Event != "USER_CREATED" {
		_ = msg.Ack(false)
		return
	}

	locale := ev.Locale
	if locale == "" {
		locale = c.cfg.DefaultLocale
	}

	if _, err := c.svc.CreateProfile(ctx, ev.UserID, &locale); err != nil {
		log.Error().Err(err).Str("user_id", ev.UserID.String()).Msg("failed to create default profile")
		_ = msg.Nack(false, true) // попробовать позже
		return
	}

	_ = msg.Ack(false)
}
