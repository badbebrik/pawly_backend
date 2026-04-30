package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/authz"
	"auth/internal/transport/http/dto"
	localemw "auth/internal/transport/http/middleware"
	"net/http"
)

func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	var req dto.ChangePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	err = h.useCases.ChangePassword.Execute(r.Context(), authuc.ChangePasswordParams{
		AccessToken: tok,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writeChangePasswordError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse())
}

func (h *Handlers) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	err := h.useCases.PasswordResetRequest.Execute(r.Context(), authuc.PasswordResetRequestParams{
		Email:  req.Email,
		Locale: localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		writePasswordResetRequestError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse())
}

func (h *Handlers) PasswordResetVerify(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetVerifyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	resp, err := h.useCases.PasswordResetVerify.Execute(r.Context(), authuc.PasswordResetVerifyParams{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		writePasswordResetVerifyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, passwordResetVerifyResponse(resp))
}

func (h *Handlers) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetConfirmRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	err := h.useCases.PasswordResetConfirm.Execute(r.Context(), authuc.PasswordResetConfirmParams{
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writePasswordResetConfirmError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse())
}
