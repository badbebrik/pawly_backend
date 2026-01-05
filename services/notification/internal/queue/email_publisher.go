package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"notification/internal/model"

	"github.com/rabbitmq/amqp091-go"
)

type EmailPublisher struct {
	ch        *amqp091.Channel
	queueName string
}

func NewEmailPublisher(ch *amqp091.Channel, queueName string) *EmailPublisher {
	return &EmailPublisher{
		ch:        ch,
		queueName: queueName,
	}
}

func (p *EmailPublisher) Publish(ctx context.Context, job model.EmailJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	if err := p.ch.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("publish email job: %w", err)
	}

	return nil
}
