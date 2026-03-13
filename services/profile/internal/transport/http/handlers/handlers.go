package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

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
	UserID            uuid.UUID `json:"user_id"`
	FirstName         *string   `json:"first_name"`
	LastName          *string   `json:"last_name"`
	AvatarDownloadURL *string   `json:"avatar_download_url"`
	Locale            string    `json:"locale"`
	Timezone          string    `json:"time_zone"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

func (h *Handlers) fromModel(ctx context.Context, p *model.Profile) ProfileResponse {
	var avatarURL *string
	if p.AvatarFileID != nil {
		if url, _, err := h.svc.GetAvatarDownloadURL(ctx, *p.AvatarFileID); err == nil {
			avatarURL = &url
		}
	}
	return ProfileResponse{
		UserID:            p.UserID,
		FirstName:         p.FirstName,
		LastName:          p.LastName,
		AvatarDownloadURL: avatarURL,
		Locale:            p.Locale,
		Timezone:          p.Timezone,
		CreatedAt:         p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.UTC().Format(time.RFC3339),
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

	writeJSON(w, http.StatusOK, h.fromModel(r.Context(), p))
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Locale    *string `json:"locale"`
	Timezone  *string `json:"time_zone"`
}

func (h *Handlers) PutMe(w http.ResponseWriter, r *http.Request) {
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
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Locale:    req.Locale,
		Timezone:  req.Timezone,
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

	writeJSON(w, http.StatusOK, h.fromModel(r.Context(), p))
}

type InitAvatarUploadRequest struct {
	MimeType          string `json:"mime_type"`
	ExpectedSizeBytes *int64 `json:"expected_size_bytes"`
}

type InitAvatarUploadResponse struct {
	FileID uuid.UUID     `json:"file_id"`
	Upload UploadInfoDTO `json:"upload"`
}

type UploadInfoDTO struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt string            `json:"expires_at"`
}

func (h *Handlers) InitAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req InitAvatarUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	fileID, upload, err := h.svc.InitAvatarUpload(r.Context(), req.MimeType, req.ExpectedSizeBytes, userID)
	if err != nil {
		http.Error(w, "init upload failed", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, InitAvatarUploadResponse{
		FileID: fileID,
		Upload: UploadInfoDTO{
			Method:    upload.Method,
			URL:       upload.URL,
			Headers:   upload.Headers,
			ExpiresAt: upload.ExpiresAt.UTC().Format(time.RFC3339),
		},
	})
}

type ConfirmAvatarUploadRequest struct {
	FileID    uuid.UUID `json:"file_id"`
	SizeBytes int64     `json:"size_bytes"`
}

type ConfirmAvatarUploadResponse struct {
	Profile ProfileResponse `json:"profile"`
}

func (h *Handlers) ConfirmAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ConfirmAvatarUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	p, err := h.svc.ConfirmAvatarUpload(r.Context(), userID, req.FileID, req.SizeBytes)
	if err != nil {
		http.Error(w, "confirm failed", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, ConfirmAvatarUploadResponse{Profile: h.fromModel(r.Context(), p)})
}

type PublicOwnerContactDTO struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
}

func (h *Handlers) GetPublicContact(w http.ResponseWriter, r *http.Request) {
	userIDRaw := chi.URLParam(r, "user_id")
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	p, err := h.svc.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var displayName *string
	name := strings.TrimSpace(strings.Join([]string{valueOrEmpty(p.FirstName), valueOrEmpty(p.LastName)}, " "))
	if name != "" {
		displayName = &name
	}

	var email *string = nil

	writeJSON(w, http.StatusOK, PublicOwnerContactDTO{
		DisplayName: displayName,
		Email:       email,
	})
}

type BatchProfilesBriefRequest struct {
	UserIDs []string `json:"user_ids"`
}

type ProfileBriefDTO struct {
	UserID            uuid.UUID `json:"user_id"`
	FirstName         *string   `json:"first_name"`
	LastName          *string   `json:"last_name"`
	DisplayName       *string   `json:"display_name"`
	AvatarDownloadURL *string   `json:"avatar_download_url"`
}

type BatchProfilesBriefResponse struct {
	Items           []ProfileBriefDTO `json:"items"`
	NotFoundUserIDs []uuid.UUID       `json:"not_found_user_ids"`
}

func (h *Handlers) BatchProfilesBrief(w http.ResponseWriter, r *http.Request) {
	var req BatchProfilesBriefRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(req.UserIDs) == 0 {
		writeJSON(w, http.StatusOK, BatchProfilesBriefResponse{
			Items:           []ProfileBriefDTO{},
			NotFoundUserIDs: []uuid.UUID{},
		})
		return
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for i := range req.UserIDs {
		userID, err := uuid.Parse(req.UserIDs[i])
		if err != nil {
			http.Error(w, "invalid user_ids", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, userID)
	}

	items, notFound, err := h.svc.BatchGetProfilesBrief(r.Context(), userIDs)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := make([]ProfileBriefDTO, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, ProfileBriefDTO{
			UserID:            item.UserID,
			FirstName:         item.FirstName,
			LastName:          item.LastName,
			DisplayName:       buildDisplayName(item.FirstName, item.LastName),
			AvatarDownloadURL: item.AvatarDownloadURL,
		})
	}

	writeJSON(w, http.StatusOK, BatchProfilesBriefResponse{
		Items:           out,
		NotFoundUserIDs: notFound,
	})
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func buildDisplayName(firstName, lastName *string) *string {
	displayName := strings.TrimSpace(strings.Join([]string{valueOrEmpty(firstName), valueOrEmpty(lastName)}, " "))
	if displayName == "" {
		return nil
	}
	return &displayName
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
