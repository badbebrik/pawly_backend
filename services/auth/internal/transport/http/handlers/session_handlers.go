package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/authz"
	"auth/internal/transport/http/dto"
	"net/http"
)

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	if err := h.useCases.Logout.Execute(r.Context(), tok); err != nil {
		writeLogoutError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.EmptyResponse{})
}

func (h *Handlers) LogoutAll(w http.ResponseWriter, r *http.Request) {
	tok, err := authz.BearerToken(r)
	if err != nil {
		writeServiceError(w, http.StatusUnauthorized, err)
		return
	}

	if err := h.useCases.LogoutAll.Execute(r.Context(), tok); err != nil {
		writeLogoutAllError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.EmptyResponse{})
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	resp, err := h.useCases.Refresh.Execute(r.Context(), authuc.RefreshParams{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeRefreshError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, sessionResponse(resp.UserID, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn))
}
