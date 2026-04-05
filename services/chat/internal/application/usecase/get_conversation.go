package usecase

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type GetConversation struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	acl           ports.ACLClient
	profiles      ports.ProfileClient
	pets          ports.PetClient
	presence      conversationPresenceReader
}

type conversationPresenceReader interface {
	IsUserInConversation(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
}

func NewGetConversation(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	acl ports.ACLClient,
	profiles ports.ProfileClient,
	pets ports.PetClient,
	presence conversationPresenceReader,
) *GetConversation {
	return &GetConversation{
		conversations: conversations,
		participants:  participants,
		acl:           acl,
		profiles:      profiles,
		pets:          pets,
		presence:      presence,
	}
}

type GetConversationParams struct {
	CurrentUserID  uuid.UUID
	ConversationID uuid.UUID
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
	ConversationID             uuid.UUID
	Pet                        GetConversationPet
	OtherUser                  GetConversationOtherUser
	LastMessageID              *uuid.UUID
	LastMessageAt              *time.Time
	LastMessagePreview         *string
	LastMessageSenderID        *uuid.UUID
	LastReadMessageID          *uuid.UUID
	OtherUserLastReadMessageID *uuid.UUID
	OtherUserInChat            bool
	UnreadCount                int
	CanSend                    bool
}

type GetConversationResult struct {
	Conversation GetConversationDetails
}

func (uc *GetConversation) Execute(ctx context.Context, params GetConversationParams) (GetConversationResult, error) {
	if params.CurrentUserID == uuid.Nil || params.ConversationID == uuid.Nil {
		return GetConversationResult{}, ErrInvalidInput
	}

	conversation, err := uc.conversations.GetByID(ctx, params.ConversationID)
	if err != nil {
		return GetConversationResult{}, err
	}
	if !conversation.HasParticipant(params.CurrentUserID) {
		return GetConversationResult{}, ErrForbidden
	}

	participant, err := uc.participants.GetByConversationAndUser(ctx, conversation.ID, params.CurrentUserID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return GetConversationResult{}, ErrForbidden
		}
		return GetConversationResult{}, err
	}

	otherUserID := conversation.OtherUserID(params.CurrentUserID)
	otherParticipant, err := uc.participants.GetByConversationAndUser(ctx, conversation.ID, otherUserID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return GetConversationResult{}, ErrForbidden
		}
		return GetConversationResult{}, err
	}

	profiles, err := uc.profiles.BatchGetBrief(ctx, []uuid.UUID{otherUserID})
	if err != nil {
		return GetConversationResult{}, err
	}

	pets, err := uc.pets.BatchGetBrief(ctx, []uuid.UUID{conversation.PetID})
	if err != nil {
		return GetConversationResult{}, err
	}

	profile, ok := profiles[otherUserID]
	if !ok {
		return GetConversationResult{}, ports.ErrNotFound
	}

	pet, ok := pets[conversation.PetID]
	if !ok {
		return GetConversationResult{}, ports.ErrNotFound
	}

	canRead, canSend, err := uc.resolveConversationAccess(ctx, conversation, params.CurrentUserID, otherUserID)
	if err != nil {
		return GetConversationResult{}, err
	}
	if !canRead {
		return GetConversationResult{}, ErrForbidden
	}

	otherUserInChat := false
	if uc.presence != nil {
		otherUserInChat, err = uc.presence.IsUserInConversation(ctx, conversation.ID, otherUserID)
		if err != nil {
			return GetConversationResult{}, err
		}
	}

	return GetConversationResult{
		Conversation: GetConversationDetails{
			ConversationID: conversation.ID,
			Pet: GetConversationPet{
				PetID:     pet.PetID,
				Name:      pet.Name,
				AvatarURL: pet.AvatarURL,
			},
			OtherUser: GetConversationOtherUser{
				UserID:      profile.UserID,
				DisplayName: profile.DisplayName,
				AvatarURL:   profile.AvatarURL,
			},
			LastMessageID:              conversation.LastMessageID,
			LastMessageAt:              conversation.LastMessageAt,
			LastMessagePreview:         conversation.LastMessagePreview,
			LastMessageSenderID:        conversation.LastMessageSenderID,
			LastReadMessageID:          participant.LastReadMessageID,
			OtherUserLastReadMessageID: otherParticipant.LastReadMessageID,
			OtherUserInChat:            otherUserInChat,
			UnreadCount:                participant.UnreadCount,
			CanSend:                    canSend,
		},
	}, nil
}

func (uc *GetConversation) resolveConversationAccess(
	ctx context.Context,
	conversation *model.Conversation,
	currentUserID, otherUserID uuid.UUID,
) (canRead bool, canSend bool, err error) {
	currentUserActive, err := uc.acl.IsActiveMember(ctx, conversation.PetID, currentUserID)
	if err != nil {
		return false, false, err
	}
	if !currentUserActive {
		return false, false, nil
	}

	otherUserActive, err := uc.acl.IsActiveMember(ctx, conversation.PetID, otherUserID)
	if err != nil {
		return false, false, err
	}

	return true, currentUserActive && otherUserActive, nil
}
