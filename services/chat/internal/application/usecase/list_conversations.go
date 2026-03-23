package usecase

import (
	"chat/internal/application/ports"
	"context"
	"time"

	"github.com/google/uuid"
)

type ListConversations struct {
	conversations ports.ConversationRepository
	profiles ports.ProfileClient
	pets     ports.PetClient
}

func NewListConversations(
	conversations ports.ConversationRepository,
	profiles ports.ProfileClient,
	pets ports.PetClient,
) *ListConversations {
	return &ListConversations{
		conversations: conversations,
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

func (uc *ListConversations) Execute(ctx context.Context, params ListConversationsParams) (ListConversationsResult, error) {
	if params.CurrentUserID == uuid.Nil {
		return ListConversationsResult{}, ErrInvalidInput
	}

	rows, err := uc.conversations.ListConversations(ctx, params.CurrentUserID, ports.ListConversationsParams{
		PetID:  params.PetID,
		Cursor: params.Cursor,
		Limit:  params.Limit,
	})
	if err != nil {
		return ListConversationsResult{}, err
	}

	petIDs := collectPetIDs(rows.Items)
	otherUserIDs := collectOtherUserIDs(rows.Items)

	petsByID, err := uc.pets.BatchGetBrief(ctx, petIDs)
	if err != nil {
		return ListConversationsResult{}, err
	}

	profilesByID, err := uc.profiles.BatchGetBrief(ctx, otherUserIDs)
	if err != nil {
		return ListConversationsResult{}, err
	}

	items := make([]ListConversationsItem, 0, len(rows.Items))
	for _, row := range rows.Items {
		pet, ok := petsByID[row.PetID]
		if !ok {
			return ListConversationsResult{}, ports.ErrNotFound
		}

		profile, ok := profilesByID[row.OtherUserID]
		if !ok {
			return ListConversationsResult{}, ports.ErrNotFound
		}

		items = append(items, ListConversationsItem{
			ConversationID: row.ConversationID,
			Pet: ListConversationsPet{
				PetID:     pet.PetID,
				Name:      pet.Name,
				AvatarURL: pet.AvatarURL,
			},
			OtherUser: ListConversationsOtherUser{
				UserID:      profile.UserID,
				DisplayName: profile.DisplayName,
				AvatarURL:   profile.AvatarURL,
			},
			LastMessageID:       row.LastMessageID,
			LastMessageAt:       row.LastMessageAt,
			LastMessagePreview:  row.LastMessagePreview,
			LastMessageSenderID: row.LastMessageSenderID,
			LastReadMessageID:   row.LastReadMessageID,
			UnreadCount:         row.UnreadCount,
		})
	}

	return ListConversationsResult{
		Items:      items,
		NextCursor: rows.NextCursor,
	}, nil
}

func collectPetIDs(rows []ports.ConversationListRow) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		if _, ok := seen[row.PetID]; ok {
			continue
		}
		seen[row.PetID] = struct{}{}
		ids = append(ids, row.PetID)
	}

	return ids
}

func collectOtherUserIDs(rows []ports.ConversationListRow) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		if _, ok := seen[row.OtherUserID]; ok {
			continue
		}
		seen[row.OtherUserID] = struct{}{}
		ids = append(ids, row.OtherUserID)
	}

	return ids
}
