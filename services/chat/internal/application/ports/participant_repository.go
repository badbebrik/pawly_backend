package ports

import (
	"chat/internal/domain/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type ParticipantRepository interface {
	CreateBatch(ctx context.Context, participants []model.ConversationParticipant) error
	GetByConversationAndUser(ctx context.Context, conversationID, userID uuid.UUID) (*model.ConversationParticipant, error)
	ListByConversation(ctx context.Context, conversationID uuid.UUID) ([]model.ConversationParticipant, error)
	IncrementUnread(ctx context.Context, conversationID, userID uuid.UUID, delta int) error
	MarkRead(ctx context.Context, conversationID, userID, lastReadMessageID uuid.UUID, readAt time.Time) (*model.ConversationParticipant, error)
}
