package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type CreateProfileInternalRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Locale *string   `json:"locale,omitempty"`
}

type createProfileInternalResponse struct {
	UserID uuid.UUID `json:"user_id"`
}

func (h *Handlers) CreateProfileInternal(w http.ResponseWriter, r *http.Request) {
	var req CreateProfileInternalRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.UserID == uuid.Nil {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	_, err := h.svc.CreateProfile(r.Context(), req.UserID, req.Locale)
	if err != nil {
		http.Error(w, "failed to create profile", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, createProfileInternalResponse{UserID: req.UserID})
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
