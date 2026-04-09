package queue

import (
	"context"
	"encoding/json"

	"health/internal/model"

	"github.com/rabbitmq/amqp091-go"
)

type PushPublisher struct {
	ch        *amqp091.Channel
	queueName string
}

func NewPushPublisher(ch *amqp091.Channel, queueName string) *PushPublisher {
	return &PushPublisher{ch: ch, queueName: queueName}
}

func (p *PushPublisher) PublishScheduledOccurrenceDue(ctx context.Context, job model.ScheduledOccurrencePushJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
