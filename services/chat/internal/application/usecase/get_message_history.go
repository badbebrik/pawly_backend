package usecase

import (
	"chat/internal/application/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type GetMessageHistory struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	messages      ports.MessageRepository
}

func NewGetMessageHistory(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	messages ports.MessageRepository,
) *GetMessageHistory {
	return &GetMessageHistory{
		conversations: conversations,
		participants:  participants,
		messages:      messages,
	}
}

type GetMessageHistoryParams struct {
	CurrentUserID    uuid.UUID
	ConversationID   uuid.UUID
	BeforeMessageID  *uuid.UUID
	Limit            int
}

type GetMessageHistoryItem struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	SenderUserID   uuid.UUID
	ClientMsgID    uuid.UUID
	Text           *string
	CreatedAt      time.Time
}

type GetMessageHistoryResult struct {
	ConversationID uuid.UUID
	Messages       []GetMessageHistoryItem
	HasMore        bool
}

func (uc *GetMessageHistory) Execute(_ context.Context, _ GetMessageHistoryParams) (GetMessageHistoryResult, error) {
	return GetMessageHistoryResult{}, ErrNotImplemented
}
