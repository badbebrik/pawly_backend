package handlers

import (
	"chat/internal/application/usecase"
	"chat/internal/transport/http/dto"
	"chat/internal/transport/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

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

	var req dto.MarkReadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	result, err := h.useCases.MarkRead.Execute(r.Context(), usecase.MarkReadParams{
		CurrentUserID:     userID,
		ConversationID:    conversationID,
		LastReadMessageID: req.LastReadMessageID,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.MarkReadResponse{
		ConversationID:    result.ConversationID,
		LastReadMessageID: result.LastReadMessageID,
	})
}
