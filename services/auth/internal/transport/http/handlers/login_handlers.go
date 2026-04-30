package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/dto"
	localemw "auth/internal/transport/http/middleware"
	"net/http"
)

func (h *Handlers) LoginEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	resp, err := h.useCases.LoginEmail.Execute(r.Context(), authuc.LoginEmailParams{
		Email:    req.Email,
		Password: req.Password,
		Locale:   localemw.LocaleFromCtx(r.Context(), "ru"),
	})
	if err != nil {
		writeLoginEmailError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sessionResponse(resp.UserID, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn))
}

func (h *Handlers) LoginOAuth(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginOAuthRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	resp, err := h.useCases.LoginOAuth.Execute(r.Context(), authuc.LoginOAuthParams{
		Provider: req.Provider,
		IDToken:  req.IDToken,
		Locale:   firstNonEmpty(req.Locale, localemw.LocaleFromCtx(r.Context(), "ru")),
		Timezone: req.Timezone,
	})
	if err != nil {
		writeLoginOAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sessionResponse(resp.UserID, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn))
}
