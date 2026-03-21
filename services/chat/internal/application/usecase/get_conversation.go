package usecase

import (
	"chat/internal/application/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type GetConversation struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	profiles      ports.ProfileClient
	pets          ports.PetClient
}

func NewGetConversation(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	profiles ports.ProfileClient,
	pets ports.PetClient,
) *GetConversation {
	return &GetConversation{
		conversations: conversations,
		participants:  participants,
		profiles:      profiles,
		pets:          pets,
	}
}

type GetConversationParams struct {
	CurrentUserID   uuid.UUID
	ConversationID  uuid.UUID
}

type GetConversationPet struct {
	PetID     uuid.UUID
	Name      string
	AvatarURL *string
}

type GetConversationOtherUser struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
}

type GetConversationDetails struct {
	ConversationID      uuid.UUID
	Pet                 GetConversationPet
	OtherUser           GetConversationOtherUser
	LastMessageID       *uuid.UUID
	LastMessageAt       *time.Time
	LastMessagePreview  *string
	LastMessageSenderID *uuid.UUID
	LastReadMessageID   *uuid.UUID
	UnreadCount         int
	CanSend             bool
}

type GetConversationResult struct {
	Conversation GetConversationDetails
}

func (uc *GetConversation) Execute(_ context.Context, _ GetConversationParams) (GetConversationResult, error) {
	return GetConversationResult{}, ErrNotImplemented
}
