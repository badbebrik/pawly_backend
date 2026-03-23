package usecase

import (
	"chat/internal/application/ports"
	"context"

	"github.com/google/uuid"
)

type GetUnreadSummary struct {
	conversations ports.ConversationRepository
}

func NewGetUnreadSummary(conversations ports.ConversationRepository) *GetUnreadSummary {
	return &GetUnreadSummary{conversations: conversations}
}

type GetUnreadSummaryParams struct {
	CurrentUserID uuid.UUID
}

type GetUnreadSummarySummary struct {
	UnreadConversations int
	UnreadMessages      int
}

type GetUnreadSummaryResult struct {
	Summary GetUnreadSummarySummary
}

func (uc *GetUnreadSummary) Execute(ctx context.Context, params GetUnreadSummaryParams) (GetUnreadSummaryResult, error) {
	if params.CurrentUserID == uuid.Nil {
		return GetUnreadSummaryResult{}, ErrInvalidInput
	}

	summary, err := uc.conversations.GetUnreadSummary(ctx, params.CurrentUserID)
	if err != nil {
		return GetUnreadSummaryResult{}, err
	}

	return GetUnreadSummaryResult{
		Summary: GetUnreadSummarySummary{
			UnreadConversations: summary.UnreadConversations,
			UnreadMessages:      summary.UnreadMessages,
		},
	}, nil
}
