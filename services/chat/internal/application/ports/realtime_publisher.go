package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MessageSentEvent struct {
	ConversationID  uuid.UUID
	MessageID       uuid.UUID
	SenderUserID    uuid.UUID
	RecipientUserID uuid.UUID
	ClientMsgID     uuid.UUID
	CreatedAt       time.Time
}

type ConversationUpdatedEvent struct {
	ConversationID uuid.UUID
	UserID         uuid.UUID
}

type GlobalUnreadUpdatedEvent struct {
	UserID              uuid.UUID
	UnreadConversations int
	UnreadMessages      int
}

type ReadUpdatedEvent struct {
	ConversationID    uuid.UUID
	UserID            uuid.UUID
	LastReadMessageID uuid.UUID
}

type RealtimePublisher interface {
	PublishMessageSent(ctx context.Context, event MessageSentEvent) error
	PublishConversationUpdated(ctx context.Context, event ConversationUpdatedEvent) error
	PublishGlobalUnreadUpdated(ctx context.Context, event GlobalUnreadUpdatedEvent) error
	PublishReadUpdated(ctx context.Context, event ReadUpdatedEvent) error
}
