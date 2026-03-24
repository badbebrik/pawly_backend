package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderUserID   uuid.UUID
	ClientMsgID    uuid.UUID
	Text           *string
	CreatedAt      time.Time
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
