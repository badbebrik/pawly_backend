package model

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID                  uuid.UUID  `json:"conversation_id"`
	PetID               uuid.UUID  `json:"pet_id"`
	UserLowID           uuid.UUID  `json:"user_low_id"`
	UserHighID          uuid.UUID  `json:"user_high_id"`
	LastMessageID       *uuid.UUID `json:"last_message_id"`
	LastMessageAt       *time.Time `json:"last_message_at"`
	LastMessagePreview  *string    `json:"last_message_preview"`
	LastMessageSenderID *uuid.UUID `json:"last_message_sender_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type ConversationParticipant struct {
	ConversationID    uuid.UUID  `json:"conversation_id"`
	UserID            uuid.UUID  `json:"user_id"`
	LastReadMessageID *uuid.UUID `json:"last_read_message_id"`
	LastReadAt        *time.Time `json:"last_read_at"`
	UnreadCount       int        `json:"unread_count"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
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
