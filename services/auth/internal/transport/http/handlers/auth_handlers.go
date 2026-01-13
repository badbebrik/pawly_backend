package handlers

import (
	authsvc "auth/internal/service"
	"auth/internal/transport/http/authz"
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

func (h *AuthHandlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.VerifyEmail(r.Context(), authsvc.VerifyEmailInput{
		Email: req.Email,
		Code:  req.Code,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("email", req.Email).
			Msg("VerifyEmail failed")
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, authsvc.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		case errors.Is(err, authsvc.ErrVerificationCodeInvalid):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		case errors.Is(err, authsvc.ErrVerificationCodeExpired),
			errors.Is(err, authsvc.ErrVerificationTooMany):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	out := dto.VerifyEmailResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AuthHandlers) LoginEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.LoginEmail(r.Context(), authsvc.LoginEmailInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("LoginEmail failed")

		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, authsvc.ErrInvalidEmailOrPassword):
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		case errors.Is(err, authsvc.ErrEmailNotVerified):
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LoginEmailResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.svc.Logout(r.Context(), tok); err != nil {
		log.Error().Err(err).Msg("Logout failed")
		switch {
		case errors.Is(err, authsvc.ErrUnauthorized):
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		case errors.Is(err, authsvc.ErrSessionNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LogoutResponse{})
}

func (h *AuthHandlers) LogoutAll(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.svc.LogoutAll(r.Context(), tok); err != nil {
		log.Error().Err(err).Msg("LogoutAll failed")
		switch {
		case errors.Is(err, authsvc.ErrUnauthorized):
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LogoutResponse{})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.Refresh(r.Context(), authsvc.RefreshInput{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		log.Error().Err(err).Msg("Refresh failed")
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, authsvc.ErrUnauthorized),
			errors.Is(err, authsvc.ErrRefreshMismatch):
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		case errors.Is(err, authsvc.ErrSessionNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		case errors.Is(err, authsvc.ErrSessionExpired),
			errors.Is(err, authsvc.ErrSessionRevoked):
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.RefreshResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
