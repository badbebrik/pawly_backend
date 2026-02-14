package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	UserID            uuid.UUID                   `json:"user_id"`
	FirstName         *string                     `json:"first_name"`
	LastName          *string                     `json:"last_name"`
	Phone             *string                     `json:"phone"`
	AvatarFileID      *uuid.UUID                  `json:"avatar_file_id"`
	AvatarDownloadURL *string                     `json:"avatar_download_url"`
	Locale            string                      `json:"locale"`
	Timezone          string                      `json:"time_zone"`
	DateFormat        string                      `json:"date_format"`
	PublicContact     model.PublicContactSettings `json:"public_contact_settings"`
	ExtraContacts     model.ExtraContacts         `json:"extra_contacts"`
	CreatedAt         string                      `json:"created_at"`
	UpdatedAt         string                      `json:"updated_at"`
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
		Phone:             p.Phone,
		AvatarFileID:      p.AvatarFileID,
		AvatarDownloadURL: avatarURL,
		Locale:            p.Locale,
		Timezone:          p.Timezone,
		DateFormat:        p.DateFormat,
		PublicContact:     p.PublicContact,
		ExtraContacts:     p.ExtraContacts,
		CreatedAt:         p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// GetMe godoc
// @Summary Get my profile
// @Tags profile
// @Produce json
// @Param X-User-ID header string true "User ID"
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/profile/me [get]
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
	FirstName     *string                      `json:"first_name"`
	LastName      *string                      `json:"last_name"`
	Phone         *string                      `json:"phone"`
	Locale        *string                      `json:"locale"`
	Timezone      *string                      `json:"time_zone"`
	DateFormat    *string                      `json:"date_format"`
	PublicContact *model.PublicContactSettings `json:"public_contact_settings"`
	ExtraContacts *model.ExtraContacts         `json:"extra_contacts"`
}

// PutMe godoc
// @Summary Update my profile
// @Tags profile
// @Accept json
// @Produce json
// @Param X-User-ID header string true "User ID"
// @Param request body UpdateProfileRequest true "request"
// @Success 200 {object} ProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/profile/me [put]
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
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Locale:        req.Locale,
		Timezone:      req.Timezone,
		DateFormat:    req.DateFormat,
		PublicContact: req.PublicContact,
		ExtraContacts: req.ExtraContacts,
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

// InitAvatarUpload godoc
// @Summary Init avatar upload
// @Tags profile
// @Accept json
// @Produce json
// @Param X-User-ID header string true "User ID"
// @Param request body InitAvatarUploadRequest true "request"
// @Success 201 {object} InitAvatarUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/profile/me/avatar:init-upload [post]
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

// ConfirmAvatarUpload godoc
// @Summary Confirm avatar upload
// @Tags profile
// @Accept json
// @Produce json
// @Param X-User-ID header string true "User ID"
// @Param request body ConfirmAvatarUploadRequest true "request"
// @Success 200 {object} ConfirmAvatarUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/profile/me/avatar:confirm-upload [post]
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

// TestAvatarUpload godoc
// @Summary Test avatar upload (multipart)
// @Tags profile
// @Accept multipart/form-data
// @Produce json
// @Param X-User-ID header string true "User ID"
// @Param file formData file true "file"
// @Success 200 {object} ConfirmAvatarUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/profile/me/avatar:test-upload [post]
// Test endpoint: accepts file in multipart/form-data and uploads via FileService presigned URL
func (h *Handlers) TestAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	size := int64(len(data))
	fileID, upload, err := h.svc.InitAvatarUpload(r.Context(), mimeType, &size, userID)
	if err != nil {
		http.Error(w, "init upload failed", http.StatusBadRequest)
		return
	}

	if err := putToURL(upload, mimeType, data); err != nil {
		http.Error(w, "upload failed", http.StatusBadRequest)
		return
	}

	p, err := h.svc.ConfirmAvatarUpload(r.Context(), userID, fileID, size)
	if err != nil {
		http.Error(w, "confirm failed", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, ConfirmAvatarUploadResponse{Profile: h.fromModel(r.Context(), p)})
}

type PublicOwnerContactDTO struct {
	DisplayName   *string             `json:"display_name"`
	Phone         *string             `json:"phone"`
	Email         *string             `json:"email"`
	ExtraContacts model.ExtraContacts `json:"extra_contacts"`
}

// GetPublicContact godoc
// @Summary Get public contact by user id
// @Tags profile
// @Produce json
// @Param user_id path string true "user id"
// @Success 200 {object} PublicOwnerContactDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /internal/v1/profile/users/{user_id}/public-contact [get]
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
	if p.PublicContact.PublicDisplayNameOverride != nil && *p.PublicContact.PublicDisplayNameOverride != "" {
		displayName = p.PublicContact.PublicDisplayNameOverride
	} else if p.PublicContact.ShowOwnerName {
		name := strings.TrimSpace(strings.Join([]string{valueOrEmpty(p.FirstName), valueOrEmpty(p.LastName)}, " "))
		if name != "" {
			displayName = &name
		}
	}

	var phone *string
	if p.PublicContact.ShowPhone {
		phone = p.Phone
	}

	var extra model.ExtraContacts = model.ExtraContacts{}
	if p.PublicContact.ShowExtraContacts {
		extra = p.ExtraContacts
	}

	// Email пока не запрашиваем из Auth
	var email *string = nil

	writeJSON(w, http.StatusOK, PublicOwnerContactDTO{
		DisplayName:   displayName,
		Phone:         phone,
		Email:         email,
		ExtraContacts: extra,
	})
}

func putToURL(upload service.UploadInfo, mimeType string, data []byte) error {
	req, err := http.NewRequest(http.MethodPut, upload.URL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mimeType)
	for k, v := range upload.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload status %d", resp.StatusCode)
	}
	return nil
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
