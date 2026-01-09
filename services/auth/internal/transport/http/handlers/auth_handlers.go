package handlers

import (
	authsvc "auth/internal/service"
	"auth/internal/transport/http/dto"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

type AuthHandlers struct {
	svc *authsvc.Service
}

func NewAuthHandlers(svc *authsvc.Service) *AuthHandlers {
	return &AuthHandlers{svc: svc}
}

func (h *AuthHandlers) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.RegisterEmail(r.Context(), authsvc.RegisterEmailInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Locale:    req.Locale,
	})

	if err != nil {
		log.Warn().Err(err).Msg("RegisterEmail failed")

		switch {
		case errors.Is(err, authsvc.ErrEmailAlreadyTaken):
			http.Error(w, err.Error(), http.StatusConflict)
			return
		case errors.Is(err, authsvc.ErrWeakPassword),
			errors.Is(err, authsvc.ErrIncorrectFormat):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, authsvc.ErrCannotResendYet):
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"error":            err.Error(),
				"user_id":          resp.UserID,
				"channel":          resp.Verification.Channel,
				"code_ttl_seconds": resp.Verification.CodeTTLSeconds,
				"can_resend_in":    resp.Verification.CanResendInSeconds,
			})
			return
		case errors.Is(err, authsvc.ErrVerificationFailed):
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	var out dto.RegisterEmailResponse
	out.UserID = resp.UserID
	out.Verification.Channel = resp.Verification.Channel
	out.Verification.CodeTTLSeconds = resp.Verification.CodeTTLSeconds
	out.Verification.CanResendInSeconds = resp.Verification.CanResendInSeconds

	writeJSON(w, http.StatusCreated, out)

}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
