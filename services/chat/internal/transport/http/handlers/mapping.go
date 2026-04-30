package handlers

import (
	chatuc "chat/internal/application/usecase"
	"chat/internal/transport/http/dto"
	"time"

	"github.com/google/uuid"
)

func openConversationToResponse(result chatuc.OpenDirectConversationResult) dto.ConversationResponse {
	return conversationToResponse(
		result.Conversation.ConversationID,
		result.Conversation.Pet.PetID,
		result.Conversation.Pet.Name,
		result.Conversation.Pet.AvatarURL,
		result.Conversation.OtherUser.UserID,
		result.Conversation.OtherUser.DisplayName,
		result.Conversation.OtherUser.AvatarURL,
		result.Conversation.LastMessageID,
		result.Conversation.LastMessageAt,
		result.Conversation.LastMessagePreview,
		result.Conversation.LastMessageSenderID,
		result.Conversation.LastReadMessageID,
		result.Conversation.OtherUserLastReadMessageID,
		result.Conversation.OtherUserInChat,
		result.Conversation.UnreadCount,
		result.Conversation.CanSend,
	)
}

func getConversationToResponse(result chatuc.GetConversationResult) dto.ConversationResponse {
	return conversationToResponse(
		result.Conversation.ConversationID,
		result.Conversation.Pet.PetID,
		result.Conversation.Pet.Name,
		result.Conversation.Pet.AvatarURL,
		result.Conversation.OtherUser.UserID,
		result.Conversation.OtherUser.DisplayName,
		result.Conversation.OtherUser.AvatarURL,
		result.Conversation.LastMessageID,
		result.Conversation.LastMessageAt,
		result.Conversation.LastMessagePreview,
		result.Conversation.LastMessageSenderID,
		result.Conversation.LastReadMessageID,
		result.Conversation.OtherUserLastReadMessageID,
		result.Conversation.OtherUserInChat,
		result.Conversation.UnreadCount,
		result.Conversation.CanSend,
	)
}

func listConversationItemToResponse(item chatuc.ListConversationsItem) dto.ConversationResponse {
	return conversationToResponse(
		item.ConversationID,
		item.Pet.PetID,
		item.Pet.Name,
		item.Pet.AvatarURL,
		item.OtherUser.UserID,
		item.OtherUser.DisplayName,
		item.OtherUser.AvatarURL,
		item.LastMessageID,
		item.LastMessageAt,
		item.LastMessagePreview,
		item.LastMessageSenderID,
		item.LastReadMessageID,
		nil,
		false,
		item.UnreadCount,
		item.CanSend,
	)
}

func conversationToResponse(
	conversationID uuid.UUID,
	petID uuid.UUID,
	petName string,
	petAvatarURL *string,
	otherUserID uuid.UUID,
	otherDisplayName *string,
	otherAvatarURL *string,
	lastMessageID *uuid.UUID,
	lastMessageAt *time.Time,
	lastMessagePreview *string,
	lastMessageSenderID *uuid.UUID,
	lastReadMessageID *uuid.UUID,
	otherUserLastReadMessageID *uuid.UUID,
	otherUserInChat bool,
	unreadCount int,
	canSend bool,
) dto.ConversationResponse {
	return dto.ConversationResponse{
		ConversationID: conversationID,
		Pet: dto.ConversationPetResponse{
			PetID:     petID,
			Name:      petName,
			AvatarURL: petAvatarURL,
		},
		OtherUser: dto.ConversationOtherUserResponse{
			UserID:      otherUserID,
			DisplayName: otherDisplayName,
			AvatarURL:   otherAvatarURL,
		},
		LastMessageID:              lastMessageID,
		LastMessageAt:              formatTimePtr(lastMessageAt),
		LastMessagePreview:         lastMessagePreview,
		LastMessageSenderID:        lastMessageSenderID,
		LastReadMessageID:          lastReadMessageID,
		OtherUserLastReadMessageID: otherUserLastReadMessageID,
		OtherUserInChat:            otherUserInChat,
		UnreadCount:                unreadCount,
		CanSend:                    canSend,
	}
}

func messageHistoryToResponse(result chatuc.GetMessageHistoryResult) dto.MessageHistoryResponse {
	items := make([]dto.MessageResponse, 0, len(result.Messages))
	for i := range result.Messages {
		items = append(items, dto.MessageResponse{
			MessageID:      result.Messages[i].MessageID,
			ConversationID: result.Messages[i].ConversationID,
			SenderUserID:   result.Messages[i].SenderUserID,
			ClientMsgID:    result.Messages[i].ClientMsgID,
			Text:           result.Messages[i].Text,
			CreatedAt:      result.Messages[i].CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return dto.MessageHistoryResponse{
		ConversationID: result.ConversationID,
		Messages:       items,
		HasMore:        result.HasMore,
	}
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
