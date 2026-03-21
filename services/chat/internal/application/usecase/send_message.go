package usecase

import (
	"chat/internal/application/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type SendMessage struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	messages      ports.MessageRepository
	tx            ports.TxManager
	acl           ports.ACLClient
	realtime      ports.RealtimePublisher
}

func NewSendMessage(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	messages ports.MessageRepository,
	tx ports.TxManager,
	acl ports.ACLClient,
	realtime ports.RealtimePublisher,
) *SendMessage {
	return &SendMessage{
		conversations: conversations,
		participants:  participants,
		messages:      messages,
		tx:            tx,
		acl:           acl,
		realtime:      realtime,
	}
}

type SendMessageParams struct {
	CurrentUserID   uuid.UUID
	ConversationID  uuid.UUID
	ClientMsgID     uuid.UUID
	Text            *string
}

type SendMessageMessage struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	SenderUserID   uuid.UUID
	ClientMsgID    uuid.UUID
	Text           *string
	CreatedAt      time.Time
}

type SendMessageResult struct {
	Message SendMessageMessage
}

func (uc *SendMessage) Execute(_ context.Context, _ SendMessageParams) (SendMessageResult, error) {
	return SendMessageResult{}, ErrNotImplemented
}
