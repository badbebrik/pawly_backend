package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"push/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	raw := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil && id != uuid.Nil {
			return id, true
		}
	}

	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" {
		return uuid.Nil, false
	}
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(auth, bearerPrefix) {
		return uuid.Nil, false
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, bearerPrefix))
	if tokenStr == "" {
		return uuid.Nil, false
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(a.cfg.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, false
	}
	if iss, ok := claims["iss"].(string); !ok || strings.TrimSpace(iss) != a.cfg.JWTIssuer {
		return uuid.Nil, false
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(strings.TrimSpace(sub))
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
