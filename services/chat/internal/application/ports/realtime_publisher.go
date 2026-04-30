package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MessageSentEvent struct {
	OriginClientID  uuid.UUID
	ConversationID  uuid.UUID
	MessageID       uuid.UUID
	SenderUserID    uuid.UUID
	RecipientUserID uuid.UUID
	ClientMsgID     uuid.UUID
	Text            *string
	CreatedAt       time.Time
}

type ReadUpdatedEvent struct {
	ConversationID    uuid.UUID
	UserID            uuid.UUID
	LastReadMessageID uuid.UUID
}

type RealtimePublisher interface {
	PublishMessageSent(ctx context.Context, event MessageSentEvent) error
	PublishReadUpdated(ctx context.Context, event ReadUpdatedEvent) error
}
