package dto

import "github.com/google/uuid"

type InitAvatarUploadRequest struct {
	MimeType          string `json:"mime_type"`
	ExpectedSizeBytes *int64 `json:"expected_size_bytes"`
}

type UploadInfoResponse struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt string            `json:"expires_at"`
}

type InitAvatarUploadResponse struct {
	FileID uuid.UUID          `json:"file_id"`
	Upload UploadInfoResponse `json:"upload"`
}

type ConfirmAvatarUploadRequest struct {
	FileID    uuid.UUID `json:"file_id"`
	SizeBytes int64     `json:"size_bytes"`
}

type ConfirmAvatarUploadResponse struct {
	Profile ProfileResponse `json:"profile"`
}
