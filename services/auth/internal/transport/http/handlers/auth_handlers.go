package handlers

import (
	authsvc "auth/internal/service"
	"auth/internal/transport/http/authz"
	"auth/internal/transport/http/dto"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type AuthHandlers struct {
	svc *authsvc.Service
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func NewAuthHandlers(svc *authsvc.Service) *AuthHandlers {
	return &AuthHandlers{svc: svc}
}

func (h *AuthHandlers) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterEmailRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "incorrect_format", nil)
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
			writeServiceError(w, http.StatusConflict, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		case errors.Is(err, authsvc.ErrWeakPassword),
			errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrCannotResendYet):
			writeError(w, http.StatusTooManyRequests, err.Error(), map[string]any{
				"user_id":          resp.UserID,
				"channel":          resp.Verification.Channel,
				"code_ttl_seconds": resp.Verification.CodeTTLSeconds,
				"can_resend_in":    resp.Verification.CanResendInSeconds,
			})
			return
		case errors.Is(err, authsvc.ErrVerificationFailed),
			errors.Is(err, authsvc.ErrProfileCreationFailed):
			writeServiceError(w, http.StatusServiceUnavailable, err)
			return
		default:
			writeInternalError(w)
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
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
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
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrUserNotFound):
			writeServiceError(w, http.StatusNotFound, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		case errors.Is(err, authsvc.ErrVerificationCodeInvalid),
			errors.Is(err, authsvc.ErrVerificationCodeExpired),
			errors.Is(err, authsvc.ErrVerificationTooMany):
			writeServiceError(w, http.StatusUnprocessableEntity, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	out := dto.VerifyEmailResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *AuthHandlers) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendEmailVerificationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.svc.ResendEmailVerification(r.Context(), authsvc.ResendEmailVerificationInput{
		Email: req.Email,
	})
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("ResendEmailVerification failed")
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrUserNotFound):
			writeServiceError(w, http.StatusNotFound, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		case errors.Is(err, authsvc.ErrEmailAlreadyVerified):
			writeServiceError(w, http.StatusConflict, err)
			return
		case errors.Is(err, authsvc.ErrCannotResendYet):
			writeError(w, http.StatusTooManyRequests, err.Error(), map[string]any{
				"user_id":          resp.UserID,
				"channel":          resp.Verification.Channel,
				"code_ttl_seconds": resp.Verification.CodeTTLSeconds,
				"can_resend_in":    resp.Verification.CanResendInSeconds,
			})
			return
		case errors.Is(err, authsvc.ErrVerificationFailed):
			writeServiceError(w, http.StatusServiceUnavailable, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	var out dto.ResendEmailVerificationResponse
	out.UserID = resp.UserID
	out.Verification.Channel = resp.Verification.Channel
	out.Verification.CodeTTLSeconds = resp.Verification.CodeTTLSeconds
	out.Verification.CanResendInSeconds = resp.Verification.CanResendInSeconds

	writeJSON(w, http.StatusOK, out)
}

func (h *AuthHandlers) LoginEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
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
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrInvalidEmailOrPassword):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrEmailNotVerified),
			errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LoginEmailResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	})
}

func (h *AuthHandlers) LoginOAuth(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginOAuthRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.svc.LoginOAuth(r.Context(), authsvc.LoginOAuthInput{
		Provider: req.Provider,
		IDToken:  req.IDToken,
	})
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Msg("LoginOAuth failed")
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrOAuthInvalidToken):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrOAuthProviderUnavailable),
			errors.Is(err, authsvc.ErrProfileCreationFailed):
			writeServiceError(w, http.StatusServiceUnavailable, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LoginOAuthResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	if err := h.svc.Logout(r.Context(), tok); err != nil {
		log.Error().Err(err).Msg("Logout failed")
		switch {
		case errors.Is(err, authsvc.ErrUnauthorized):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrSessionNotFound):
			writeServiceError(w, http.StatusNotFound, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LogoutResponse{})
}

func (h *AuthHandlers) LogoutAll(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	if err := h.svc.LogoutAll(r.Context(), tok); err != nil {
		log.Error().Err(err).Msg("LogoutAll failed")
		switch {
		case errors.Is(err, authsvc.ErrUnauthorized):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.LogoutResponse{})
}

func (h *AuthHandlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	var req dto.ChangePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	err = h.svc.ChangePassword(r.Context(), authsvc.ChangePasswordInput{
		AccessToken: tok,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		log.Error().Err(err).Msg("ChangePassword failed")
		switch {
		case errors.Is(err, authsvc.ErrUnauthorized):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrInvalidEmailOrPassword):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrIncorrectFormat),
			errors.Is(err, authsvc.ErrWeakPassword):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrUserNotFound):
			writeServiceError(w, http.StatusNotFound, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.ChangePasswordResponse{Status: "ok"})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.svc.Refresh(r.Context(), authsvc.RefreshInput{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		log.Error().Err(err).Msg("Refresh failed")
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrUnauthorized),
			errors.Is(err, authsvc.ErrRefreshMismatch),
			errors.Is(err, authsvc.ErrSessionExpired),
			errors.Is(err, authsvc.ErrSessionRevoked):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrSessionNotFound):
			writeServiceError(w, http.StatusNotFound, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.RefreshResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	})
}

func (h *AuthHandlers) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequestRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	err := h.svc.RequestPasswordReset(r.Context(), authsvc.PasswordResetRequestInput{Email: req.Email})
	if err != nil {
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrCannotResendYet):
			writeServiceError(w, http.StatusTooManyRequests, err)
			return
		case errors.Is(err, authsvc.ErrVerificationFailed):
			writeServiceError(w, http.StatusServiceUnavailable, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.PasswordResetRequestResponse{Status: "ok"})
}

func (h *AuthHandlers) PasswordResetVerify(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetVerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.svc.VerifyPasswordResetCode(r.Context(), authsvc.PasswordResetVerifyInput{Email: req.Email, Code: req.Code})
	if err != nil {
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrVerificationCodeInvalid),
			errors.Is(err, authsvc.ErrVerificationCodeExpired),
			errors.Is(err, authsvc.ErrVerificationTooMany):
			writeServiceError(w, http.StatusUnprocessableEntity, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.PasswordResetVerifyResponse{ResetToken: resp.ResetToken})
}

func (h *AuthHandlers) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetConfirmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	err := h.svc.ConfirmPasswordReset(r.Context(), authsvc.PasswordResetConfirmInput{
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, authsvc.ErrIncorrectFormat):
			writeServiceError(w, http.StatusBadRequest, err)
			return
		case errors.Is(err, authsvc.ErrUnauthorized):
			writeServiceError(w, http.StatusUnauthorized, err)
			return
		case errors.Is(err, authsvc.ErrUserNotFound):
			writeServiceError(w, http.StatusNotFound, err)
			return
		case errors.Is(err, authsvc.ErrUserBlocked):
			writeServiceError(w, http.StatusForbidden, err)
			return
		default:
			writeInternalError(w)
			return
		}
	}

	writeJSON(w, http.StatusOK, dto.PasswordResetConfirmResponse{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeServiceError(w http.ResponseWriter, status int, err error) {
	writeError(w, status, err.Error(), nil)
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "internal_error", nil)
}

func writeError(w http.ResponseWriter, status int, code string, details any) {
	writeJSON(w, status, errorResponse{
		Code:    code,
		Message: code,
		Details: details,
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
