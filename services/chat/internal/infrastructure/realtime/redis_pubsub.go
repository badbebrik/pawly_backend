package realtime

import (
	"chat/internal/application/usecase"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	EventTypeMessageSent                 = "message_sent"
	EventTypeReadUpdated                 = "read_updated"
	EventTypeConversationPresenceUpdated = "conversation_presence_updated"
)

type EventPublisher interface {
	PublishMessageSent(ctx context.Context, event MessageSentEvent) error
	PublishReadUpdated(ctx context.Context, event ReadUpdatedEvent) error
	PublishConversationPresenceUpdated(ctx context.Context, event ConversationPresenceUpdatedEvent) error
}

type MessageSentEvent struct {
	OriginClientID  uuid.UUID `json:"origin_client_id"`
	ConversationID  uuid.UUID `json:"conversation_id"`
	MessageID       uuid.UUID `json:"message_id"`
	SenderUserID    uuid.UUID `json:"sender_user_id"`
	RecipientUserID uuid.UUID `json:"recipient_user_id"`
	ClientMsgID     uuid.UUID `json:"client_msg_id"`
	Text            *string   `json:"text,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReadUpdatedEvent struct {
	ConversationID    uuid.UUID `json:"conversation_id"`
	UserID            uuid.UUID `json:"user_id"`
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

type ConversationPresenceUpdatedEvent struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	IsInChat       bool      `json:"is_in_chat"`
}

type eventEnvelope struct {
	Type                        string                            `json:"type"`
	MessageSent                 *MessageSentEvent                 `json:"message_sent,omitempty"`
	ReadUpdated                 *ReadUpdatedEvent                 `json:"read_updated,omitempty"`
	ConversationPresenceUpdated *ConversationPresenceUpdatedEvent `json:"conversation_presence_updated,omitempty"`
}

type RedisPublisher struct {
	client  *redis.Client
	channel string
}

func NewRedisPublisher(client *redis.Client, channel string) *RedisPublisher {
	return &RedisPublisher{client: client, channel: channel}
}

func (p *RedisPublisher) PublishMessageSent(ctx context.Context, event MessageSentEvent) error {
	return p.publish(ctx, eventEnvelope{Type: EventTypeMessageSent, MessageSent: &event})
}

func (p *RedisPublisher) PublishReadUpdated(ctx context.Context, event ReadUpdatedEvent) error {
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
	client           *redis.Client
	channel          string
	hub              *Hub
	getConversation  *usecase.GetConversation
	getUnreadSummary *usecase.GetUnreadSummary
}

func NewRedisSubscriber(
	client *redis.Client,
	channel string,
	hub *Hub,
	getConversation *usecase.GetConversation,
	getUnreadSummary *usecase.GetUnreadSummary,
) *RedisSubscriber {
	return &RedisSubscriber{
		client:           client,
		channel:          channel,
		hub:              hub,
		getConversation:  getConversation,
		getUnreadSummary: getUnreadSummary,
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
		if event.MessageSent != nil {
			s.handleMessageSent(ctx, *event.MessageSent)
		}
	case EventTypeReadUpdated:
		if event.ReadUpdated != nil {
			s.handleReadUpdated(ctx, *event.ReadUpdated)
		}
	case EventTypeConversationPresenceUpdated:
		if event.ConversationPresenceUpdated != nil {
			s.handleConversationPresenceUpdated(*event.ConversationPresenceUpdated)
		}
	}
}

func (s *RedisSubscriber) handleMessageSent(ctx context.Context, event MessageSentEvent) {
	if s.hub.HasConversationSubscribers(event.ConversationID) {
		payload := mustMarshal(outboundEnvelope{
			Type: "message_new",
			Payload: messageNewPayload{
				MessageID:      event.MessageID,
				ConversationID: event.ConversationID,
				SenderUserID:   event.SenderUserID,
				ClientMsgID:    event.ClientMsgID,
				Text:           event.Text,
				CreatedAt:      event.CreatedAt.UTC().Format(time.RFC3339),
			},
		})
		if len(payload) > 0 {
			s.hub.PublishToConversationExceptClientID(event.ConversationID, event.OriginClientID, payload)
		}
	}

	s.publishConversationInboxIfSubscribed(ctx, event.SenderUserID, event.ConversationID)
	s.publishGlobalUnreadIfSubscribed(ctx, event.SenderUserID)
	s.publishConversationInboxIfSubscribed(ctx, event.RecipientUserID, event.ConversationID)
	s.publishGlobalUnreadIfSubscribed(ctx, event.RecipientUserID)
}

func (s *RedisSubscriber) handleReadUpdated(ctx context.Context, event ReadUpdatedEvent) {
	if s.hub.HasConversationSubscribers(event.ConversationID) {
		payload := mustMarshal(outboundEnvelope{
			Type: "read_updated",
			Payload: readUpdatedPayload{
				ConversationID:    event.ConversationID,
				UserID:            event.UserID,
				LastReadMessageID: event.LastReadMessageID,
			},
		})
		if len(payload) > 0 {
			s.hub.PublishToConversation(event.ConversationID, payload)
		}
	}

	s.publishConversationInboxIfSubscribed(ctx, event.UserID, event.ConversationID)
	s.publishGlobalUnreadIfSubscribed(ctx, event.UserID)
}

func (s *RedisSubscriber) handleConversationPresenceUpdated(event ConversationPresenceUpdatedEvent) {
	if !s.hub.HasConversationSubscribers(event.ConversationID) {
		return
	}

	payload := mustMarshal(outboundEnvelope{
		Type: "conversation_presence_updated",
		Payload: conversationPresenceUpdatedPayload{
			ConversationID: event.ConversationID,
			UserID:         event.UserID,
			IsInChat:       event.IsInChat,
		},
	})
	if len(payload) > 0 {
		s.hub.PublishToConversation(event.ConversationID, payload)
	}
}

func (s *RedisSubscriber) publishConversationInboxIfSubscribed(ctx context.Context, userID, conversationID uuid.UUID) {
	if !s.hub.HasUserInboxSubscribers(userID) {
		return
	}

	result, err := s.getConversation.Execute(ctx, usecase.GetConversationParams{
		CurrentUserID:  userID,
		ConversationID: conversationID,
	})
	if err != nil {
		return
	}

	payload := mustMarshal(outboundEnvelope{
		Type: "conversation_updated",
		Payload: conversationUpdatedPayload{
			ConversationID: result.Conversation.ConversationID,
			Pet: conversationPetPayload{
				PetID:     result.Conversation.Pet.PetID,
				Name:      result.Conversation.Pet.Name,
				AvatarURL: result.Conversation.Pet.AvatarURL,
			},
			OtherUser: conversationOtherUserPayload{
				UserID:      result.Conversation.OtherUser.UserID,
				DisplayName: result.Conversation.OtherUser.DisplayName,
				AvatarURL:   result.Conversation.OtherUser.AvatarURL,
			},
			LastMessageID:              result.Conversation.LastMessageID,
			LastMessageAt:              formatTimePtr(result.Conversation.LastMessageAt),
			LastMessagePreview:         result.Conversation.LastMessagePreview,
			LastMessageSenderID:        result.Conversation.LastMessageSenderID,
			LastReadMessageID:          result.Conversation.LastReadMessageID,
			OtherUserLastReadMessageID: result.Conversation.OtherUserLastReadMessageID,
			OtherUserInChat:            result.Conversation.OtherUserInChat,
			UnreadCount:                result.Conversation.UnreadCount,
			CanSend:                    result.Conversation.CanSend,
		},
	})
	if len(payload) > 0 {
		s.hub.PublishToUserInbox(userID, payload)
	}
}

func (s *RedisSubscriber) publishGlobalUnreadIfSubscribed(ctx context.Context, userID uuid.UUID) {
	if !s.hub.HasUserInboxSubscribers(userID) {
		return
	}

	result, err := s.getUnreadSummary.Execute(ctx, usecase.GetUnreadSummaryParams{
		CurrentUserID: userID,
	})
	if err != nil {
		return
	}

	payload := mustMarshal(outboundEnvelope{
		Type: "global_unread_updated",
		Payload: globalUnreadUpdatedPayload{
			UnreadConversations: result.UnreadConversations,
			UnreadMessages:      result.UnreadMessages,
		},
	})
	if len(payload) > 0 {
		s.hub.PublishToUserInbox(userID, payload)
	}
}

func mustMarshal(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}

type outboundEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type messageNewPayload struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

type readUpdatedPayload struct {
	ConversationID    uuid.UUID `json:"conversation_id"`
	UserID            uuid.UUID `json:"user_id"`
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

type conversationPetPayload struct {
	PetID     uuid.UUID `json:"pet_id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

type conversationOtherUserPayload struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName *string   `json:"display_name,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

type conversationUpdatedPayload struct {
	ConversationID             uuid.UUID                    `json:"conversation_id"`
	Pet                        conversationPetPayload       `json:"pet"`
	OtherUser                  conversationOtherUserPayload `json:"other_user"`
	LastMessageID              *uuid.UUID                   `json:"last_message_id,omitempty"`
	LastMessageAt              *string                      `json:"last_message_at,omitempty"`
	LastMessagePreview         *string                      `json:"last_message_preview,omitempty"`
	LastMessageSenderID        *uuid.UUID                   `json:"last_message_sender_id,omitempty"`
	LastReadMessageID          *uuid.UUID                   `json:"last_read_message_id,omitempty"`
	OtherUserLastReadMessageID *uuid.UUID                   `json:"other_user_last_read_message_id,omitempty"`
	OtherUserInChat            bool                         `json:"other_user_in_chat"`
	UnreadCount                int                          `json:"unread_count"`
	CanSend                    bool                         `json:"can_send"`
}

type conversationPresenceUpdatedPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	IsInChat       bool      `json:"is_in_chat"`
}

type globalUnreadUpdatedPayload struct {
	UnreadConversations int `json:"unread_conversations"`
	UnreadMessages      int `json:"unread_messages"`
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
