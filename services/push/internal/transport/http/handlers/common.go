package handlers

import (
	"errors"
	"net/http"
	"strings"

	pushuc "push/internal/application/usecase"

	"github.com/google/uuid"
	"pawly/pkg/gatewayauth"
	"pawly/pkg/httpjson"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	httpjson.Write(w, status, payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpjson.WriteError(w, status, code, message)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pushuc.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, pushuc.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, pushuc.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "conflict")
	case errors.Is(err, pushuc.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid input")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func decodeRequest(r *http.Request, dst any) error {
	return httpjson.Decode(r, dst)
}

func requireUserID(r *http.Request) (uuid.UUID, bool) {
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
