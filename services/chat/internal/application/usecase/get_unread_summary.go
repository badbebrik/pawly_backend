package usecase

import (
	"chat/internal/application/ports"
	"context"

	"github.com/google/uuid"
)

type GetUnreadSummary struct {
	read ports.ConversationReadRepository
}

func NewGetUnreadSummary(read ports.ConversationReadRepository) *GetUnreadSummary {
	return &GetUnreadSummary{read: read}
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

func (uc *GetUnreadSummary) Execute(_ context.Context, _ GetUnreadSummaryParams) (GetUnreadSummaryResult, error) {
	return GetUnreadSummaryResult{}, ErrNotImplemented
}
