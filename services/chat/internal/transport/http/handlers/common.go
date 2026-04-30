package handlers

import (
	"chat/internal/application/ports"
	chatuc "chat/internal/application/usecase"
	"errors"
	"net/http"
	"pawly/pkg/httpjson"

	"github.com/google/uuid"
)

type Handlers struct {
	useCases *chatuc.Set
}

func New(useCases *chatuc.Set) *Handlers {
	return &Handlers{useCases: useCases}
}

func decodeJSON(r *http.Request, dst any) error {
	return httpjson.Decode(r, dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	httpjson.Write(w, status, payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpjson.WriteError(w, status, code, message)
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
