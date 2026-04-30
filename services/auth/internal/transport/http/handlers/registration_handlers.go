package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/dto"
	localemw "auth/internal/transport/http/middleware"
	"net/http"
)

func (h *Handlers) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "incorrect_format", "incorrect format")
		return
	}

	resp, err := h.useCases.RegisterEmail.Execute(r.Context(), authuc.RegisterEmailParams{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Locale:    firstNonEmpty(req.Locale, localemw.LocaleFromCtx(r.Context(), "ru")),
		Timezone:  req.Timezone,
	})
	if err != nil {
		writeRegisterEmailError(w, err, resp)
		return
	}

	writeJSON(w, http.StatusCreated, registerEmailResponse(resp))
}

func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	resp, err := h.useCases.VerifyEmail.Execute(r.Context(), authuc.VerifyEmailParams{
		Email:  req.Email,
		Code:   req.Code,
		Locale: localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		writeVerifyEmailError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sessionResponse(resp.UserID, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn))
}

func (h *Handlers) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendEmailVerificationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	resp, err := h.useCases.ResendEmailVerification.Execute(r.Context(), authuc.ResendEmailVerificationParams{
		Email:  req.Email,
		Locale: localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		writeResendEmailVerificationError(w, err, resp)
		return
	}

	writeJSON(w, http.StatusOK, resendEmailVerificationResponse(resp))
}
