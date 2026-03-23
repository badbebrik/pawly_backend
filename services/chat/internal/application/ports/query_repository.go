package ports

import (
	"time"

	"github.com/google/uuid"
)

type ListConversationsParams struct {
	PetID  *uuid.UUID
	Cursor *string
	Limit  int
}

type ConversationListRow struct {
	ConversationID       uuid.UUID
	PetID                uuid.UUID
	OtherUserID          uuid.UUID
	LastMessageID        *uuid.UUID
	LastMessageAt        *time.Time
	LastMessagePreview   *string
	LastMessageSenderID  *uuid.UUID
	LastReadMessageID    *uuid.UUID
	UnreadCount          int
}

type ListConversationsResult struct {
	Items      []ConversationListRow
	NextCursor *string
}

type UnreadSummary struct {
	UnreadConversations int
	UnreadMessages      int
}
