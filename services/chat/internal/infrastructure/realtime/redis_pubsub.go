package realtime

import (
	"chat/internal/application/ports"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	EventTypeMessageSent                 = "message_sent"
	EventTypeReadUpdated                 = "read_updated"
	EventTypeConversationPresenceUpdated = "conversation_presence_updated"
)

type EventPublisher interface {
	PublishConversationPresenceUpdated(ctx context.Context, event ConversationPresenceUpdatedEvent) error
}

type SubscriberHandler interface {
	HandleMessageSent(ctx context.Context, event ports.MessageSentEvent)
	HandleReadUpdated(ctx context.Context, event ports.ReadUpdatedEvent)
	HandleConversationPresenceUpdated(event ConversationPresenceUpdatedEvent)
}

type ConversationPresenceUpdatedEvent struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	IsInChat       bool      `json:"is_in_chat"`
}

type eventEnvelope struct {
	Type                        string                            `json:"type"`
	MessageSent                 *ports.MessageSentEvent           `json:"message_sent,omitempty"`
	ReadUpdated                 *ports.ReadUpdatedEvent           `json:"read_updated,omitempty"`
	ConversationPresenceUpdated *ConversationPresenceUpdatedEvent `json:"conversation_presence_updated,omitempty"`
}

type RedisPublisher struct {
	client  *redis.Client
	channel string
}

func NewRedisPublisher(client *redis.Client, channel string) *RedisPublisher {
	return &RedisPublisher{client: client, channel: channel}
}

func (p *RedisPublisher) PublishMessageSent(ctx context.Context, event ports.MessageSentEvent) error {
	return p.publish(ctx, eventEnvelope{Type: EventTypeMessageSent, MessageSent: &event})
}

func (p *RedisPublisher) PublishReadUpdated(ctx context.Context, event ports.ReadUpdatedEvent) error {
	return p.publish(ctx, eventEnvelope{Type: EventTypeReadUpdated, ReadUpdated: &event})
}

func (p *RedisPublisher) PublishConversationPresenceUpdated(ctx context.Context, event ConversationPresenceUpdatedEvent) error {
	return p.publish(ctx, eventEnvelope{
		Type:                        EventTypeConversationPresenceUpdated,
		ConversationPresenceUpdated: &event,
	})
}

func (p *RedisPublisher) publish(ctx context.Context, event eventEnvelope) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, p.channel, payload).Err()
}

type RedisSubscriber struct {
	client  *redis.Client
	channel string
	handler SubscriberHandler
}

func NewRedisSubscriber(
	client *redis.Client,
	channel string,
	handler SubscriberHandler,
) *RedisSubscriber {
	return &RedisSubscriber{
		client:  client,
		channel: channel,
		handler: handler,
	}
}

func (s *RedisSubscriber) Run(ctx context.Context) error {
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer func() { _ = pubsub.Close() }()

	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			s.handleMessage(ctx, msg.Payload)
		}
	}
}

func (s *RedisSubscriber) handleMessage(ctx context.Context, raw string) {
	var event eventEnvelope
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return
	}

	switch event.Type {
	case EventTypeMessageSent:
		if event.MessageSent != nil && s.handler != nil {
			s.handler.HandleMessageSent(ctx, *event.MessageSent)
		}
	case EventTypeReadUpdated:
		if event.ReadUpdated != nil && s.handler != nil {
			s.handler.HandleReadUpdated(ctx, *event.ReadUpdated)
		}
	case EventTypeConversationPresenceUpdated:
		if event.ConversationPresenceUpdated != nil && s.handler != nil {
			s.handler.HandleConversationPresenceUpdated(*event.ConversationPresenceUpdated)
		}
	}
}
