package queue

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Handle(ctx context.Context, msg amqp091.Delivery)
}

type Consumer struct {
	ch        *amqp091.Channel
	queueName string
	handler   Handler
}

func NewConsumer(ch *amqp091.Channel, queueName string, handler Handler) *Consumer {
	return &Consumer{
		ch:        ch,
		queueName: queueName,
		handler:   handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.ch.Consume(
		c.queueName,
		"",
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
				c.handler.Handle(ctx, msg)
			}
		}
	}()

	return nil
}
