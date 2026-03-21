package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

func (m *Message) Normalize() {
	if m.Text == nil {
		return
	}

	trimmed := strings.TrimSpace(*m.Text)
	if trimmed == "" {
		m.Text = nil
		return
	}

	m.Text = &trimmed
}

func (m *Message) ValidateForSend() error {
	if m.ClientMsgID == uuid.Nil {
		return ErrMessageClientIDRequired
	}

	m.Normalize()
	if m.Text == nil {
		return ErrMessageTextRequired
	}

	return nil
}
