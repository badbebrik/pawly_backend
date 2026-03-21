package ports

import (
	"chat/internal/domain/model"
	"context"

	"github.com/google/uuid"
)

type MessageHistoryPage struct {
	Messages []model.Message
	HasMore  bool
}

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	GetByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error)
	FindByClientMsgID(ctx context.Context, conversationID, senderUserID, clientMsgID uuid.UUID) (*model.Message, error)
	ListHistory(ctx context.Context, conversationID uuid.UUID, beforeMessageID *uuid.UUID, limit int) (MessageHistoryPage, error)
}
