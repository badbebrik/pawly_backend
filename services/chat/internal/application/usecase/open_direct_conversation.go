package usecase

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type OpenDirectConversation struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	tx            ports.TxManager
	acl           ports.ACLClient
	profiles      ports.ProfileClient
	pets          ports.PetClient
	presence      conversationPresenceReader
}

func NewOpenDirectConversation(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	tx ports.TxManager,
	acl ports.ACLClient,
	profiles ports.ProfileClient,
	pets ports.PetClient,
	presence conversationPresenceReader,
) *OpenDirectConversation {
	return &OpenDirectConversation{
		conversations: conversations,
		participants:  participants,
		tx:            tx,
		acl:           acl,
		profiles:      profiles,
		pets:          pets,
		presence:      presence,
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
	ConversationID             uuid.UUID
	Pet                        OpenDirectConversationPet
	OtherUser                  OpenDirectConversationOtherUser
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

type OpenDirectConversationResult struct {
	Conversation OpenDirectConversationConversation
}

func (uc *OpenDirectConversation) Execute(ctx context.Context, params OpenDirectConversationParams) (OpenDirectConversationResult, error) {
	if params.CurrentUserID == uuid.Nil || params.PetID == uuid.Nil || params.OtherUserID == uuid.Nil {
		return OpenDirectConversationResult{}, ErrInvalidInput
	}
	if params.CurrentUserID == params.OtherUserID {
		return OpenDirectConversationResult{}, ErrInvalidInput
	}

	userLowID, userHighID, err := model.NormalizeDirectUserPair(params.CurrentUserID, params.OtherUserID)
	if err != nil {
		return OpenDirectConversationResult{}, ErrInvalidInput
	}

	conversation, err := uc.conversations.GetDirectByPetAndUsers(ctx, params.PetID, userLowID, userHighID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return OpenDirectConversationResult{}, err
	}

	if conversation != nil {
		participant, err := uc.participants.GetByConversationAndUser(ctx, conversation.ID, params.CurrentUserID)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				return OpenDirectConversationResult{}, ErrForbidden
			}
			return OpenDirectConversationResult{}, err
		}

		return uc.buildResult(ctx, conversation, participant, params.CurrentUserID)
	}

	currentUserActive, err := uc.acl.IsActiveMember(ctx, params.PetID, params.CurrentUserID)
	if err != nil {
		return OpenDirectConversationResult{}, err
	}
	if !currentUserActive {
		return OpenDirectConversationResult{}, ErrForbidden
	}

	otherUserActive, err := uc.acl.IsActiveMember(ctx, params.PetID, params.OtherUserID)
	if err != nil {
		return OpenDirectConversationResult{}, err
	}
	if !otherUserActive {
		return OpenDirectConversationResult{}, ErrForbidden
	}

	now := time.Now().UTC()
	conversation = &model.Conversation{
		ID:         uuid.New(),
		PetID:      params.PetID,
		UserLowID:  userLowID,
		UserHighID: userHighID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	participants := []model.ConversationParticipant{
		{
			ConversationID: conversation.ID,
			UserID:         params.CurrentUserID,
			UnreadCount:    0,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ConversationID: conversation.ID,
			UserID:         params.OtherUserID,
			UnreadCount:    0,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	err = uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := uc.conversations.Create(ctx, conversation); err != nil {
			return err
		}
		if err := uc.participants.CreateBatch(ctx, participants); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if !errors.Is(err, ports.ErrConflict) {
			return OpenDirectConversationResult{}, err
		}

		conversation, err = uc.conversations.GetDirectByPetAndUsers(ctx, params.PetID, userLowID, userHighID)
		if err != nil {
			return OpenDirectConversationResult{}, err
		}
	}

	participant, err := uc.participants.GetByConversationAndUser(ctx, conversation.ID, params.CurrentUserID)
	if err != nil {
		return OpenDirectConversationResult{}, err
	}

	return uc.buildResult(ctx, conversation, participant, params.CurrentUserID)
}

func (uc *OpenDirectConversation) buildResult(
	ctx context.Context,
	conversation *model.Conversation,
	participant *model.ConversationParticipant,
	currentUserID uuid.UUID,
) (OpenDirectConversationResult, error) {
	if !conversation.HasParticipant(currentUserID) {
		return OpenDirectConversationResult{}, ErrForbidden
	}

	otherUserID := conversation.OtherUserID(currentUserID)
	otherParticipant, err := uc.participants.GetByConversationAndUser(ctx, conversation.ID, otherUserID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return OpenDirectConversationResult{}, ErrForbidden
		}
		return OpenDirectConversationResult{}, err
	}

	profiles, err := uc.profiles.BatchGetBrief(ctx, []uuid.UUID{otherUserID})
	if err != nil {
		return OpenDirectConversationResult{}, err
	}

	pets, err := uc.pets.BatchGetBrief(ctx, []uuid.UUID{conversation.PetID})
	if err != nil {
		return OpenDirectConversationResult{}, err
	}

	profile, ok := profiles[otherUserID]
	if !ok {
		return OpenDirectConversationResult{}, ports.ErrNotFound
	}

	pet, ok := pets[conversation.PetID]
	if !ok {
		return OpenDirectConversationResult{}, ports.ErrNotFound
	}

	currentUserActive, err := uc.acl.IsActiveMember(ctx, conversation.PetID, currentUserID)
	if err != nil {
		return OpenDirectConversationResult{}, err
	}
	if !currentUserActive {
		return OpenDirectConversationResult{}, ErrForbidden
	}

	otherUserActive, err := uc.acl.IsActiveMember(ctx, conversation.PetID, otherUserID)
	if err != nil {
		return OpenDirectConversationResult{}, err
	}

	otherUserInChat := false
	if uc.presence != nil {
		otherUserInChat, err = uc.presence.IsUserInConversation(ctx, conversation.ID, otherUserID)
		if err != nil {
			return OpenDirectConversationResult{}, err
		}
	}

	return OpenDirectConversationResult{
		Conversation: OpenDirectConversationConversation{
			ConversationID: conversation.ID,
			Pet: OpenDirectConversationPet{
				PetID:     pet.PetID,
				Name:      pet.Name,
				AvatarURL: pet.AvatarURL,
			},
			OtherUser: OpenDirectConversationOtherUser{
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
			CanSend:                    currentUserActive && otherUserActive,
		},
	}, nil
}
