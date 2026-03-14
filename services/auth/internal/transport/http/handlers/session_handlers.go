package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/authz"
	"auth/internal/transport/http/dto"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	if err := h.uc.Logout.Execute(r.Context(), tok); err != nil {
		log.Error().Err(err).Msg("Logout failed")
		writeLogoutError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.LogoutResponse{})
}

func (h *AuthHandlers) LogoutAll(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	if err := h.uc.LogoutAll.Execute(r.Context(), tok); err != nil {
		log.Error().Err(err).Msg("LogoutAll failed")
		writeLogoutAllError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.LogoutResponse{})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	resp, err := h.uc.Refresh.Execute(r.Context(), authuc.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		log.Error().Err(err).Msg("Refresh failed")
		writeRefreshError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.RefreshResponse{
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    resp.ExpiresIn,
	})
}
