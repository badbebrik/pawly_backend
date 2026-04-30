package ws

import (
	"chat/internal/application/ports"
	"chat/internal/application/usecase"
	rtinfra "chat/internal/infrastructure/realtime"
	"chat/internal/realtime"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SubscriberHandler struct {
	hub              *realtime.Hub
	getConversation  *usecase.GetConversation
	getUnreadSummary *usecase.GetUnreadSummary
}

func NewSubscriberHandler(
	hub *realtime.Hub,
	getConversation *usecase.GetConversation,
	getUnreadSummary *usecase.GetUnreadSummary,
) *SubscriberHandler {
	return &SubscriberHandler{
		hub:              hub,
		getConversation:  getConversation,
		getUnreadSummary: getUnreadSummary,
	}
}

func (h *SubscriberHandler) HandleMessageSent(ctx context.Context, event ports.MessageSentEvent) {
	if h.hub.HasConversationSubscribers(event.ConversationID) {
		payload := marshalOutbound(outboundEnvelope{
			Type: "message_new",
			Payload: messageNewPayload{
				MessageID:      event.MessageID,
				ConversationID: event.ConversationID,
				SenderUserID:   event.SenderUserID,
				ClientMsgID:    event.ClientMsgID,
				Text:           event.Text,
				CreatedAt:      event.CreatedAt.UTC().Format(time.RFC3339),
			},
		})
		if len(payload) > 0 {
			h.hub.PublishToConversationExceptClientID(event.ConversationID, event.OriginClientID, payload)
		}
	}

	h.publishConversationInboxIfSubscribed(ctx, event.SenderUserID, event.ConversationID)
	h.publishGlobalUnreadIfSubscribed(ctx, event.SenderUserID)
	h.publishConversationInboxIfSubscribed(ctx, event.RecipientUserID, event.ConversationID)
	h.publishGlobalUnreadIfSubscribed(ctx, event.RecipientUserID)
}

func (h *SubscriberHandler) HandleReadUpdated(ctx context.Context, event ports.ReadUpdatedEvent) {
	if h.hub.HasConversationSubscribers(event.ConversationID) {
		payload := marshalOutbound(outboundEnvelope{
			Type: "read_updated",
			Payload: readUpdatedPayload{
				ConversationID:    event.ConversationID,
				UserID:            event.UserID,
				LastReadMessageID: event.LastReadMessageID,
			},
		})
		if len(payload) > 0 {
			h.hub.PublishToConversation(event.ConversationID, payload)
		}
	}

	h.publishConversationInboxIfSubscribed(ctx, event.UserID, event.ConversationID)
	h.publishGlobalUnreadIfSubscribed(ctx, event.UserID)
}

func (h *SubscriberHandler) HandleConversationPresenceUpdated(event rtinfra.ConversationPresenceUpdatedEvent) {
	if !h.hub.HasConversationSubscribers(event.ConversationID) {
		return
	}

	payload := marshalOutbound(outboundEnvelope{
		Type: "conversation_presence_updated",
		Payload: conversationPresenceUpdatedPayload{
			ConversationID: event.ConversationID,
			UserID:         event.UserID,
			IsInChat:       event.IsInChat,
		},
	})
	if len(payload) > 0 {
		h.hub.PublishToConversation(event.ConversationID, payload)
	}
}

func (h *SubscriberHandler) publishConversationInboxIfSubscribed(ctx context.Context, userID, conversationID uuid.UUID) {
	if !h.hub.HasUserInboxSubscribers(userID) {
		return
	}

	result, err := h.getConversation.Execute(ctx, usecase.GetConversationParams{
		CurrentUserID:  userID,
		ConversationID: conversationID,
	})
	if err != nil {
		return
	}

	payload := marshalOutbound(outboundEnvelope{
		Type: "conversation_updated",
		Payload: conversationUpdatedPayload{
			ConversationID: result.Conversation.ConversationID,
			Pet: conversationPetPayload{
				PetID:     result.Conversation.Pet.PetID,
				Name:      result.Conversation.Pet.Name,
				AvatarURL: result.Conversation.Pet.AvatarURL,
			},
			OtherUser: conversationOtherUserPayload{
				UserID:      result.Conversation.OtherUser.UserID,
				DisplayName: result.Conversation.OtherUser.DisplayName,
				AvatarURL:   result.Conversation.OtherUser.AvatarURL,
			},
			LastMessageID:              result.Conversation.LastMessageID,
			LastMessageAt:              formatTimePtr(result.Conversation.LastMessageAt),
			LastMessagePreview:         result.Conversation.LastMessagePreview,
			LastMessageSenderID:        result.Conversation.LastMessageSenderID,
			LastReadMessageID:          result.Conversation.LastReadMessageID,
			OtherUserLastReadMessageID: result.Conversation.OtherUserLastReadMessageID,
			OtherUserInChat:            result.Conversation.OtherUserInChat,
			UnreadCount:                result.Conversation.UnreadCount,
			CanSend:                    result.Conversation.CanSend,
		},
	})
	if len(payload) > 0 {
		h.hub.PublishToUserInbox(userID, payload)
	}
}

func (h *SubscriberHandler) publishGlobalUnreadIfSubscribed(ctx context.Context, userID uuid.UUID) {
	if !h.hub.HasUserInboxSubscribers(userID) {
		return
	}

	result, err := h.getUnreadSummary.Execute(ctx, usecase.GetUnreadSummaryParams{
		CurrentUserID: userID,
	})
	if err != nil {
		return
	}

	payload := marshalOutbound(outboundEnvelope{
		Type: "global_unread_updated",
		Payload: globalUnreadUpdatedPayload{
			UnreadConversations: result.UnreadConversations,
			UnreadMessages:      result.UnreadMessages,
		},
	})
	if len(payload) > 0 {
		h.hub.PublishToUserInbox(userID, payload)
	}
}

func marshalOutbound(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}
