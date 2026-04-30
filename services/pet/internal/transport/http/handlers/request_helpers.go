package handlers

import (
	"net/http"
	appmw "pet/internal/transport/http/middleware"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func requireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return uuid.Nil, false
	}

	return userID, true
}

func decodeRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := decodeJSON(r, dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return false
	}

	return true
}

func parseRouteUUID(w http.ResponseWriter, r *http.Request, param, field string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid "+field)
		return uuid.Nil, false
	}

	return id, true
}

func parseOptionalDate(raw *string, layout, field string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	value, err := time.Parse(layout, *raw)
	if err != nil {
		return nil, invalidFieldError{field: field}
	}

	return &value, nil
}

type invalidFieldError struct {
	field string
}

func (e invalidFieldError) Error() string {
	return "invalid " + e.field
}
