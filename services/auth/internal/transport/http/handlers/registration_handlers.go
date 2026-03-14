package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/dto"
	localemw "auth/internal/transport/http/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

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

	resp, err := h.uc.RegisterEmail.Execute(r.Context(), authuc.RegisterEmailInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Locale:    firstNonEmpty(req.Locale, localemw.LocaleFromCtx(r.Context(), "ru")),
		Timezone:  req.Timezone,
	})
	if err != nil {
		log.Warn().Err(err).Msg("RegisterEmail failed")
		writeRegisterEmailError(w, err, resp)
		return
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

	resp, err := h.uc.VerifyEmail.Execute(r.Context(), authuc.VerifyEmailInput{
		Email:  req.Email,
		Code:   req.Code,
		Locale: localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("VerifyEmail failed")
		writeVerifyEmailError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.VerifyEmailResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	})
}

func (h *AuthHandlers) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendEmailVerificationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.uc.ResendEmailVerification.Execute(r.Context(), authuc.ResendEmailVerificationInput{
		Email:  req.Email,
		Locale: localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("ResendEmailVerification failed")
		writeResendEmailVerificationError(w, err, resp)
		return
	}

	var out dto.ResendEmailVerificationResponse
	out.UserID = resp.UserID
	out.Verification.Channel = resp.Verification.Channel
	out.Verification.CodeTTLSeconds = resp.Verification.CodeTTLSeconds
	out.Verification.CanResendInSeconds = resp.Verification.CanResendInSeconds

	writeJSON(w, http.StatusOK, out)
}
