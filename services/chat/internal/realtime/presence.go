package realtime

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PresenceChange struct {
	ConversationID uuid.UUID
	UserID         uuid.UUID
	IsInChat       bool
}

type PresenceReader interface {
	IsUserInConversation(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
}

type PresenceTracker interface {
	PresenceReader
	SetInConversation(ctx context.Context, conversationID, userID, clientID uuid.UUID, ttl time.Duration) (PresenceChange, bool, error)
	RefreshConversations(ctx context.Context, userID, clientID uuid.UUID, conversationIDs []uuid.UUID, ttl time.Duration) error
	ClearInConversation(ctx context.Context, conversationID, userID, clientID uuid.UUID) (PresenceChange, bool, error)
	ClearClient(ctx context.Context, userID, clientID uuid.UUID, conversationIDs []uuid.UUID) ([]PresenceChange, error)
}
