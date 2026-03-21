package usecase

import (
	"chat/internal/application/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type ListConversations struct {
	read     ports.ConversationReadRepository
	profiles ports.ProfileClient
	pets     ports.PetClient
}

func NewListConversations(
	read ports.ConversationReadRepository,
	profiles ports.ProfileClient,
	pets ports.PetClient,
) *ListConversations {
	return &ListConversations{
		read:     read,
		profiles: profiles,
		pets:     pets,
	}
}

type ListConversationsParams struct {
	CurrentUserID uuid.UUID
	PetID         *uuid.UUID
	Cursor        *string
	Limit         int
}

type ListConversationsPet struct {
	PetID     uuid.UUID
	Name      string
	AvatarURL *string
}

type ListConversationsOtherUser struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
}

type ListConversationsItem struct {
	ConversationID      uuid.UUID
	Pet                 ListConversationsPet
	OtherUser           ListConversationsOtherUser
	LastMessageID       *uuid.UUID
	LastMessageAt       *time.Time
	LastMessagePreview  *string
	LastMessageSenderID *uuid.UUID
	LastReadMessageID   *uuid.UUID
	UnreadCount         int
}

type ListConversationsResult struct {
	Items      []ListConversationsItem
	NextCursor *string
}

func (uc *ListConversations) Execute(_ context.Context, _ ListConversationsParams) (ListConversationsResult, error) {
	return ListConversationsResult{}, ErrNotImplemented
}
