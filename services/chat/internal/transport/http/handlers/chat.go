package handlers

import (
	chatuc "chat/internal/application/usecase"
	"chat/internal/transport/http/middleware"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type openConversationRequest struct {
	PetID       uuid.UUID `json:"pet_id"`
	OtherUserID uuid.UUID `json:"other_user_id"`
}

type conversationPetResponse struct {
	PetID     uuid.UUID `json:"pet_id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

type conversationOtherUserResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName *string   `json:"display_name,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

type conversationResponse struct {
	ConversationID      uuid.UUID                       `json:"conversation_id"`
	Pet                 conversationPetResponse         `json:"pet"`
	OtherUser           conversationOtherUserResponse   `json:"other_user"`
	LastMessageID       *uuid.UUID                      `json:"last_message_id,omitempty"`
	LastMessageAt       *string                         `json:"last_message_at,omitempty"`
	LastMessagePreview  *string                         `json:"last_message_preview,omitempty"`
	LastMessageSenderID *uuid.UUID                      `json:"last_message_sender_id,omitempty"`
	LastReadMessageID   *uuid.UUID                      `json:"last_read_message_id,omitempty"`
	UnreadCount         int                             `json:"unread_count"`
	CanSend             bool                            `json:"can_send"`
}

type listConversationsResponse struct {
	Items      []conversationResponse `json:"items"`
	NextCursor *string                `json:"next_cursor,omitempty"`
}

type unreadSummaryResponse struct {
	UnreadConversations int `json:"unread_conversations"`
	UnreadMessages      int `json:"unread_messages"`
}

type messageResponse struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

type messageHistoryResponse struct {
	ConversationID uuid.UUID          `json:"conversation_id"`
	Messages       []messageResponse  `json:"messages"`
	HasMore        bool               `json:"has_more"`
}

type sendMessageRequest struct {
	ClientMsgID uuid.UUID `json:"client_msg_id"`
	Text        *string   `json:"text"`
}

type markReadRequest struct {
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

type markReadResponse struct {
	ConversationID    uuid.UUID `json:"conversation_id"`
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

func (h *Handlers) OpenConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req openConversationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	result, err := h.openDirectConversation.Execute(r.Context(), chatuc.OpenDirectConversationParams{
		CurrentUserID: userID,
		PetID:         req.PetID,
		OtherUserID:   req.OtherUserID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapOpenConversationResponse(result))
}

func (h *Handlers) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var petID *uuid.UUID
	if rawPetID := r.URL.Query().Get("pet_id"); rawPetID != "" {
		id, err := parseUUID(rawPetID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_pet_id", "invalid pet_id")
			return
		}
		petID = &id
	}

	var limit int
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_limit", "invalid limit")
			return
		}
		limit = parsed
	}

	var cursor *string
	if rawCursor := r.URL.Query().Get("cursor"); rawCursor != "" {
		cursor = &rawCursor
	}

	result, err := h.listConversations.Execute(r.Context(), chatuc.ListConversationsParams{
		CurrentUserID: userID,
		PetID:         petID,
		Cursor:        cursor,
		Limit:         limit,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	items := make([]conversationResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, mapListConversationItem(item))
	}

	writeJSON(w, http.StatusOK, listConversationsResponse{
		Items:      items,
		NextCursor: result.NextCursor,
	})
}

func (h *Handlers) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	conversationID, err := parseUUID(chi.URLParam(r, "conversation_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_conversation_id", "invalid conversation_id")
		return
	}

	result, err := h.getConversation.Execute(r.Context(), chatuc.GetConversationParams{
		CurrentUserID:  userID,
		ConversationID: conversationID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapGetConversationResponse(result))
}

func (h *Handlers) GetUnreadSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	result, err := h.getUnreadSummary.Execute(r.Context(), chatuc.GetUnreadSummaryParams{
		CurrentUserID: userID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, unreadSummaryResponse{
		UnreadConversations: result.Summary.UnreadConversations,
		UnreadMessages:      result.Summary.UnreadMessages,
	})
}

func (h *Handlers) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	conversationID, err := parseUUID(chi.URLParam(r, "conversation_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_conversation_id", "invalid conversation_id")
		return
	}

	var beforeMessageID *uuid.UUID
	if rawBefore := r.URL.Query().Get("before_message_id"); rawBefore != "" {
		id, err := parseUUID(rawBefore)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_before_message_id", "invalid before_message_id")
			return
		}
		beforeMessageID = &id
	}

	var limit int
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_limit", "invalid limit")
			return
		}
		limit = parsed
	}

	result, err := h.getMessageHistory.Execute(r.Context(), chatuc.GetMessageHistoryParams{
		CurrentUserID:   userID,
		ConversationID:  conversationID,
		BeforeMessageID: beforeMessageID,
		Limit:           limit,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	items := make([]messageResponse, 0, len(result.Messages))
	for _, item := range result.Messages {
		items = append(items, messageResponse{
			MessageID:      item.MessageID,
			ConversationID: item.ConversationID,
			SenderUserID:   item.SenderUserID,
			ClientMsgID:    item.ClientMsgID,
			Text:           item.Text,
			CreatedAt:      item.CreatedAt.UTC().Format(http.TimeFormat),
		})
	}

	writeJSON(w, http.StatusOK, messageHistoryResponse{
		ConversationID: result.ConversationID,
		Messages:       items,
		HasMore:        result.HasMore,
	})
}

func (h *Handlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	conversationID, err := parseUUID(chi.URLParam(r, "conversation_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_conversation_id", "invalid conversation_id")
		return
	}

	var req sendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	result, err := h.sendMessage.Execute(r.Context(), chatuc.SendMessageParams{
		CurrentUserID:  userID,
		ConversationID: conversationID,
		ClientMsgID:    req.ClientMsgID,
		Text:           req.Text,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, messageResponse{
		MessageID:      result.Message.MessageID,
		ConversationID: result.Message.ConversationID,
		SenderUserID:   result.Message.SenderUserID,
		ClientMsgID:    result.Message.ClientMsgID,
		Text:           result.Message.Text,
		CreatedAt:      result.Message.CreatedAt.UTC().Format(http.TimeFormat),
	})
}

func (h *Handlers) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	conversationID, err := parseUUID(chi.URLParam(r, "conversation_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_conversation_id", "invalid conversation_id")
		return
	}

	var req markReadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	result, err := h.markRead.Execute(r.Context(), chatuc.MarkReadParams{
		CurrentUserID:     userID,
		ConversationID:    conversationID,
		LastReadMessageID: req.LastReadMessageID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, markReadResponse{
		ConversationID:    result.ConversationID,
		LastReadMessageID: result.LastReadMessageID,
	})
}

func mapOpenConversationResponse(result chatuc.OpenDirectConversationResult) conversationResponse {
	conversation := result.Conversation
	return conversationResponse{
		ConversationID:      conversation.ConversationID,
		Pet:                 mapConversationPet(conversation.Pet.PetID, conversation.Pet.Name, conversation.Pet.AvatarURL),
		OtherUser:           mapOtherUser(conversation.OtherUser.UserID, conversation.OtherUser.DisplayName, conversation.OtherUser.AvatarURL),
		LastMessageID:       conversation.LastMessageID,
		LastMessageAt:       formatTimePtr(conversation.LastMessageAt),
		LastMessagePreview:  conversation.LastMessagePreview,
		LastMessageSenderID: conversation.LastMessageSenderID,
		LastReadMessageID:   conversation.LastReadMessageID,
		UnreadCount:         conversation.UnreadCount,
		CanSend:             conversation.CanSend,
	}
}

func mapGetConversationResponse(result chatuc.GetConversationResult) conversationResponse {
	conversation := result.Conversation
	return conversationResponse{
		ConversationID:      conversation.ConversationID,
		Pet:                 mapConversationPet(conversation.Pet.PetID, conversation.Pet.Name, conversation.Pet.AvatarURL),
		OtherUser:           mapOtherUser(conversation.OtherUser.UserID, conversation.OtherUser.DisplayName, conversation.OtherUser.AvatarURL),
		LastMessageID:       conversation.LastMessageID,
		LastMessageAt:       formatTimePtr(conversation.LastMessageAt),
		LastMessagePreview:  conversation.LastMessagePreview,
		LastMessageSenderID: conversation.LastMessageSenderID,
		LastReadMessageID:   conversation.LastReadMessageID,
		UnreadCount:         conversation.UnreadCount,
		CanSend:             conversation.CanSend,
	}
}

func mapListConversationItem(item chatuc.ListConversationsItem) conversationResponse {
	return conversationResponse{
		ConversationID:      item.ConversationID,
		Pet:                 mapConversationPet(item.Pet.PetID, item.Pet.Name, item.Pet.AvatarURL),
		OtherUser:           mapOtherUser(item.OtherUser.UserID, item.OtherUser.DisplayName, item.OtherUser.AvatarURL),
		LastMessageID:       item.LastMessageID,
		LastMessageAt:       formatTimePtr(item.LastMessageAt),
		LastMessagePreview:  item.LastMessagePreview,
		LastMessageSenderID: item.LastMessageSenderID,
		LastReadMessageID:   item.LastReadMessageID,
		UnreadCount:         item.UnreadCount,
	}
}

func mapConversationPet(petID uuid.UUID, name string, avatarURL *string) conversationPetResponse {
	return conversationPetResponse{
		PetID:     petID,
		Name:      name,
		AvatarURL: avatarURL,
	}
}

func mapOtherUser(userID uuid.UUID, displayName, avatarURL *string) conversationOtherUserResponse {
	return conversationOtherUserResponse{
		UserID:      userID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
	}
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
