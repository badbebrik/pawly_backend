package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
	"profile/internal/model"
	"profile/internal/repository"
	"profile/internal/service"
	"profile/internal/transport/http/middleware"
)

type Handlers struct {
	svc *service.ProfileService
}

func NewHandlers(svc *service.ProfileService) *Handlers {
	return &Handlers{svc: svc}
}

type ProfileResponse struct {
	UserID        uuid.UUID      `json:"user_id"`
	FirstName     *string        `json:"first_name"`
	LastName      *string        `json:"last_name"`
	AvatarURL     *string        `json:"avatar_url"`
	Phone         *string        `json:"phone"`
	Locale        string         `json:"locale"`
	Timezone      string         `json:"timezone"`
	DateFormat    string         `json:"date_format"`
	Notifications map[string]any `json:"notifications"`
}

func fromModel(p *model.Profile) ProfileResponse {
	return ProfileResponse{
		UserID:        p.UserID,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		AvatarURL:     p.AvatarURL,
		Phone:         p.Phone,
		Locale:        p.Locale,
		Timezone:      p.Timezone,
		DateFormat:    p.DateFormat,
		Notifications: p.Notifications,
	}
}

func (h *Handlers) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	p, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("GetProfile failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, fromModel(p))
}

type UpdateProfileRequest struct {
	FirstName     *string        `json:"first_name"`
	LastName      *string        `json:"last_name"`
	AvatarURL     *string        `json:"avatar_url"`
	Phone         *string        `json:"phone"`
	Locale        *string        `json:"locale"`
	Timezone      *string        `json:"timezone"`
	DateFormat    *string        `json:"date_format"`
	Notifications map[string]any `json:"notifications"`
}

func (h *Handlers) PatchMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	in := service.UpdateProfileInput{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		AvatarURL:     req.AvatarURL,
		Phone:         req.Phone,
		Locale:        req.Locale,
		Timezone:      req.Timezone,
		DateFormat:    req.DateFormat,
		Notifications: req.Notifications,
	}

	p, err := h.svc.UpdateProfile(r.Context(), userID, in)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("UpdateProfile failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, fromModel(p))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
