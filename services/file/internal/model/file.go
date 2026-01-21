package model

import (
	"github.com/google/uuid"
	"time"
)

type FileStatus string

const (
	FileStatusUploading FileStatus = "UPLOADING"
	FileStatusReady     FileStatus = "READY"
	FileStatusFailed    FileStatus = "FAILED"
	FileStatusDeleted   FileStatus = "DELETED"
)

type OwnerService string

const (
	OwnerServiceProfile OwnerService = "PROFILE"
	OwnerServicePet     OwnerService = "PET"
	OwnerServiceGuide   OwnerService = "GUIDE"
	OwnerServiceLog     OwnerService = "LOG"
	OwnerServiceHealth  OwnerService = "HEALTH"
	OwnerServiceChat    OwnerService = "CHAT"
	OwnerServiceCatalog OwnerService = "CATALOG"
)

type FileObject struct {
	ID              uuid.UUID  `json:"id"`
	Bucket          string     `json:"bucket"`
	ObjectKey       string     `json:"object_key"`
	Status          FileStatus `json:"status"`
	MimeType        string     `json:"mime_type"`
	SizeBytes       *int64     `json:"size_bytes"`
	CreatedByUserID uuid.UUID  `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UploadExpiresAt time.Time  `json:"upload_expires_at"`
}

type FileLink struct {
	ID              uuid.UUID   `json:"id"`
	FileID          uuid.UUID   `json:"file_id"`
	OwnerService    OwnerService `json:"owner_service"`
	OwnerType       string      `json:"owner_type"`
	OwnerID         uuid.UUID   `json:"owner_id"`
	PetID           *uuid.UUID  `json:"pet_id"`
	CreatedByUserID uuid.UUID   `json:"created_by_user_id"`
	CreatedAt       time.Time   `json:"created_at"`
}
