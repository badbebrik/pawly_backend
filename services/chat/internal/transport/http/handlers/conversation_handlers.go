package handlers

import (
	"chat/internal/application/usecase"
	"chat/internal/transport/http/dto"
	"chat/internal/transport/http/middleware"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) OpenConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req dto.OpenConversationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	result, err := h.useCases.OpenDirectConversation.Execute(r.Context(), usecase.OpenDirectConversationParams{
		CurrentUserID: userID,
		PetID:         req.PetID,
		OtherUserID:   req.OtherUserID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, openConversationToResponse(result))
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

	result, err := h.useCases.ListConversations.Execute(r.Context(), usecase.ListConversationsParams{
		CurrentUserID: userID,
		PetID:         petID,
		Cursor:        cursor,
		Limit:         limit,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	items := make([]dto.ConversationResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, listConversationItemToResponse(result.Items[i]))
	}

	writeJSON(w, http.StatusOK, dto.ListConversationsResponse{
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

	result, err := h.useCases.GetConversation.Execute(r.Context(), usecase.GetConversationParams{
		CurrentUserID:  userID,
		ConversationID: conversationID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, getConversationToResponse(result))
}

func (h *Handlers) GetUnreadSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	result, err := h.useCases.GetUnreadSummary.Execute(r.Context(), usecase.GetUnreadSummaryParams{
		CurrentUserID: userID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.UnreadSummaryResponse{
		UnreadConversations: result.UnreadConversations,
		UnreadMessages:      result.UnreadMessages,
	})
}
