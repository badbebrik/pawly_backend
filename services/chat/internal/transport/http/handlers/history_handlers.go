package handlers

import (
	"chat/internal/application/usecase"
	"chat/internal/transport/http/middleware"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

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

	result, err := h.useCases.GetMessageHistory.Execute(r.Context(), usecase.GetMessageHistoryParams{
		CurrentUserID:   userID,
		ConversationID:  conversationID,
		BeforeMessageID: beforeMessageID,
		Limit:           limit,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, messageHistoryToResponse(result))
}
