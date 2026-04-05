package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	profileuc "profile/internal/application/usecase"
	"profile/internal/model"
	"profile/internal/transport/http/middleware"
)

type Handlers struct {
	uc *profileuc.Set
}

func NewHandlers(uc *profileuc.Set) *Handlers {
	return &Handlers{uc: uc}
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
		if url, err := h.uc.GetAvatarDownloadURL.Execute(ctx, *p.AvatarFileID); err == nil {
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
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	p, err := h.uc.GetProfile.Execute(r.Context(), userID)
	if err != nil {
		writeProfileQueryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, h.fromModel(r.Context(), p))
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

func (h *Handlers) PatchMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	in := profileuc.UpdateProfileInfoInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	p, err := h.uc.UpdateProfileInfo.Execute(r.Context(), userID, in)
	if err != nil {
		writeUpdateProfileInfoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, h.fromModel(r.Context(), p))
}

type UpdatePreferencesRequest struct {
	Locale   *string `json:"locale"`
	Timezone *string `json:"time_zone"`
}

func (h *Handlers) PatchPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req UpdatePreferencesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	p, err := h.uc.UpdatePreferences.Execute(r.Context(), userID, profileuc.UpdatePreferencesInput{
		Locale:   req.Locale,
		Timezone: req.Timezone,
	})
	if err != nil {
		writeUpdatePreferencesError(w, err)
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
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req InitAvatarUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	fileID, upload, err := h.uc.InitAvatarUpload.Execute(r.Context(), userID, req.MimeType, req.ExpectedSizeBytes)
	if err != nil {
		writeAvatarError(w, err)
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
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var req ConfirmAvatarUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}

	p, err := h.uc.ConfirmAvatarUpload.Execute(r.Context(), userID, req.FileID, req.SizeBytes)
	if err != nil {
		writeAvatarError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ConfirmAvatarUploadResponse{Profile: h.fromModel(r.Context(), p)})
}

func (h *Handlers) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	p, err := h.uc.DeleteAvatar.Execute(r.Context(), userID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ConfirmAvatarUploadResponse{Profile: h.fromModel(r.Context(), p)})
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
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", nil)
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
			writeError(w, http.StatusBadRequest, "invalid_user_ids", nil)
			return
		}
		userIDs = append(userIDs, userID)
	}

	items, notFound, err := h.uc.BatchProfilesBrief.Execute(r.Context(), userIDs)
	if err != nil {
		writeBatchProfilesBriefError(w, err)
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
