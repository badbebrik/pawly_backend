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

func (uc *GetMessageHistory) Execute(ctx context.Context, params GetMessageHistoryParams) (GetMessageHistoryResult, error) {
	if params.CurrentUserID == uuid.Nil || params.ConversationID == uuid.Nil {
		return GetMessageHistoryResult{}, ErrInvalidInput
	}

	conversation, err := uc.conversations.GetByID(ctx, params.ConversationID)
	if err != nil {
		return GetMessageHistoryResult{}, err
	}
	if !conversation.HasParticipant(params.CurrentUserID) {
		return GetMessageHistoryResult{}, ErrForbidden
	}

	_, err = uc.participants.GetByConversationAndUser(ctx, params.ConversationID, params.CurrentUserID)
	if err != nil {
		if err == ports.ErrNotFound {
			return GetMessageHistoryResult{}, ErrForbidden
		}
		return GetMessageHistoryResult{}, err
	}

	page, err := uc.messages.ListHistory(ctx, params.ConversationID, params.BeforeMessageID, params.Limit)
	if err != nil {
		return GetMessageHistoryResult{}, err
	}

	items := make([]GetMessageHistoryItem, 0, len(page.Messages))
	for _, message := range page.Messages {
		items = append(items, GetMessageHistoryItem{
			MessageID:      message.ID,
			ConversationID: message.ConversationID,
			SenderUserID:   message.SenderUserID,
			ClientMsgID:    message.ClientMsgID,
			Text:           message.Text,
			CreatedAt:      message.CreatedAt,
		})
	}

	return GetMessageHistoryResult{
		ConversationID: params.ConversationID,
		Messages:       items,
		HasMore:        page.HasMore,
	}, nil
}
