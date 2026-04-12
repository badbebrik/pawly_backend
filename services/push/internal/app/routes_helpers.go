package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"push/internal/service"

	"github.com/google/uuid"
	"pawly/pkg/gatewayauth"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "conflict")
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "validation failed")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}
}

func (a *App) getUserID(r *http.Request) (uuid.UUID, bool) {
	id, err := gatewayauth.UserIDFromRequestHeader(r)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

func parsePetID(path string) (uuid.UUID, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "pets" || parts[3] != "push-settings" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func parseDeviceID(path string) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "push" || parts[2] != "devices" {
		return "", false
	}
	deviceID := strings.TrimSpace(parts[3])
	if deviceID == "" {
		return "", false
	}
	return deviceID, true
}
