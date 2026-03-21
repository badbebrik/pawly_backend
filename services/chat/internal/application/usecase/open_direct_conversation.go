package usecase

import (
	"chat/internal/application/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type OpenDirectConversation struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	tx            ports.TxManager
	acl           ports.ACLClient
}

func NewOpenDirectConversation(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	tx ports.TxManager,
	acl ports.ACLClient,
) *OpenDirectConversation {
	return &OpenDirectConversation{
		conversations: conversations,
		participants:  participants,
		tx:            tx,
		acl:           acl,
	}
}

type OpenDirectConversationParams struct {
	CurrentUserID uuid.UUID
	PetID         uuid.UUID
	OtherUserID   uuid.UUID
}

type OpenDirectConversationPet struct {
	PetID     uuid.UUID
	Name      string
	AvatarURL *string
}

type OpenDirectConversationOtherUser struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
}

type OpenDirectConversationConversation struct {
	ConversationID      uuid.UUID
	Pet                 OpenDirectConversationPet
	OtherUser           OpenDirectConversationOtherUser
	LastMessageID       *uuid.UUID
	LastMessageAt       *time.Time
	LastMessagePreview  *string
	LastMessageSenderID *uuid.UUID
	LastReadMessageID   *uuid.UUID
	UnreadCount         int
}

type OpenDirectConversationResult struct {
	Conversation OpenDirectConversationConversation
}

func (uc *OpenDirectConversation) Execute(_ context.Context, _ OpenDirectConversationParams) (OpenDirectConversationResult, error) {
	return OpenDirectConversationResult{}, ErrNotImplemented
}
