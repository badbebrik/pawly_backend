package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/authz"
	"auth/internal/transport/http/dto"
	localemw "auth/internal/transport/http/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

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

	err = h.uc.ChangePassword.Execute(r.Context(), authuc.ChangePasswordInput{
		AccessToken: tok,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		log.Error().Err(err).Msg("ChangePassword failed")
		writeChangePasswordError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ChangePasswordResponse{Status: "ok"})
}

func (h *AuthHandlers) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequestRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	err := h.uc.PasswordResetRequest.Execute(r.Context(), authuc.PasswordResetRequestInput{
		Email:  req.Email,
		Locale: localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		writePasswordResetRequestError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PasswordResetRequestResponse{Status: "ok"})
}

func (h *AuthHandlers) PasswordResetVerify(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetVerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.uc.PasswordResetVerify.Execute(r.Context(), authuc.PasswordResetVerifyInput{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		writePasswordResetVerifyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PasswordResetVerifyResponse{ResetToken: resp.ResetToken})
}

func (h *AuthHandlers) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetConfirmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	err := h.uc.PasswordResetConfirm.Execute(r.Context(), authuc.PasswordResetConfirmInput{
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writePasswordResetConfirmError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.PasswordResetConfirmResponse{Status: "ok"})
}
