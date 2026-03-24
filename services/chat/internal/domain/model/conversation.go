package model

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	UserLowID           uuid.UUID
	UserHighID          uuid.UUID
	LastMessageID       *uuid.UUID
	LastMessageAt       *time.Time
	LastMessagePreview  *string
	LastMessageSenderID *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ConversationParticipant struct {
	ConversationID    uuid.UUID
	UserID            uuid.UUID
	LastReadMessageID *uuid.UUID
	LastReadAt        *time.Time
	UnreadCount       int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NormalizeDirectUserPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	if a == uuid.Nil || b == uuid.Nil || a == b {
		return uuid.Nil, uuid.Nil, ErrConversationUserPairInvalid
	}

	if a.String() < b.String() {
		return a, b, nil
	}

	return b, a, nil
}

func (c *Conversation) HasParticipant(userID uuid.UUID) bool {
	return c.UserLowID == userID || c.UserHighID == userID
}

func (c *Conversation) OtherUserID(userID uuid.UUID) uuid.UUID {
	if c.UserLowID == userID {
		return c.UserHighID
	}

	return c.UserLowID
}
