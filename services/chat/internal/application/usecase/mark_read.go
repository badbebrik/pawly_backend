package usecase

import (
	"chat/internal/application/ports"
	"context"

	"github.com/google/uuid"
)

type MarkRead struct {
	participants ports.ParticipantRepository
	tx           ports.TxManager
	realtime     ports.RealtimePublisher
}

func NewMarkRead(
	participants ports.ParticipantRepository,
	tx ports.TxManager,
	realtime ports.RealtimePublisher,
) *MarkRead {
	return &MarkRead{
		participants: participants,
		tx:           tx,
		realtime:     realtime,
	}
}

type MarkReadParams struct {
	CurrentUserID     uuid.UUID
	ConversationID    uuid.UUID
	LastReadMessageID uuid.UUID
}

type MarkReadResult struct {
	ConversationID    uuid.UUID
	LastReadMessageID uuid.UUID
}

func (uc *MarkRead) Execute(_ context.Context, _ MarkReadParams) (MarkReadResult, error) {
	return MarkReadResult{}, ErrNotImplemented
}
