package handlers

import (
	"chat/internal/application/ports"
	chatuc "chat/internal/application/usecase"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type Handlers struct {
	openDirectConversation *chatuc.OpenDirectConversation
	listConversations      *chatuc.ListConversations
	getConversation        *chatuc.GetConversation
	getUnreadSummary       *chatuc.GetUnreadSummary
	getMessageHistory      *chatuc.GetMessageHistory
	sendMessage            *chatuc.SendMessage
	markRead               *chatuc.MarkRead
}

func New(
	openDirectConversation *chatuc.OpenDirectConversation,
	listConversations *chatuc.ListConversations,
	getConversation *chatuc.GetConversation,
	getUnreadSummary *chatuc.GetUnreadSummary,
	getMessageHistory *chatuc.GetMessageHistory,
	sendMessage *chatuc.SendMessage,
	markRead *chatuc.MarkRead,
) *Handlers {
	return &Handlers{
		openDirectConversation: openDirectConversation,
		listConversations:      listConversations,
		getConversation:        getConversation,
		getUnreadSummary:       getUnreadSummary,
		getMessageHistory:      getMessageHistory,
		sendMessage:            sendMessage,
		markRead:               markRead,
	}
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("invalid_json")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{
		Code:    code,
		Message: message,
	})
}

func writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, chatuc.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid input")
	case errors.Is(err, chatuc.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, ports.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, ports.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func parseUUID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}
