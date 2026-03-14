package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/dto"
	localemw "auth/internal/transport/http/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *AuthHandlers) LoginEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.uc.LoginEmail.Execute(r.Context(), authuc.LoginEmailInput{
		Email:    req.Email,
		Password: req.Password,
		Locale:   localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("LoginEmail failed")
		writeLoginEmailError(w, err)
		return
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

	resp, err := h.uc.LoginOAuth.Execute(r.Context(), authuc.LoginOAuthInput{
		Provider: req.Provider,
		IDToken:  req.IDToken,
		Locale:   firstNonEmpty(req.Locale, localemw.LocaleFromCtx(r.Context(), "ru")),
		Timezone: req.Timezone,
	})
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Msg("LoginOAuth failed")
		writeLoginOAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.LoginOAuthResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	})
}
