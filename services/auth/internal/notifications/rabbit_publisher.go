package notifications

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	ch    *amqp091.Channel
	queue string
}

func NewRabbitPublisher(ch *amqp091.Channel, queue string) *RabbitPublisher {
	return &RabbitPublisher{
		ch:    ch,
		queue: queue,
	}
}

func (p *RabbitPublisher) publish(ctx context.Context, ev NotificationEvent) error {
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(
		ctx,
		"",
		p.queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
