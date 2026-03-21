package ports

import (
	"chat/internal/domain/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type ConversationRepository interface {
	Create(ctx context.Context, conversation *model.Conversation) error
	GetByID(ctx context.Context, conversationID uuid.UUID) (*model.Conversation, error)
	GetDirectByPetAndUsers(ctx context.Context, petID, userLowID, userHighID uuid.UUID) (*model.Conversation, error)
	UpdateLastMessage(ctx context.Context, conversationID, messageID, senderUserID uuid.UUID, preview *string, createdAt time.Time) error
}
