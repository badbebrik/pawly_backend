package model

import (
	"github.com/google/uuid"
	"time"
)

type FileStatus string

const (
	FileStatusUploading     FileStatus = "UPLOADING"
	FileStatusReady         FileStatus = "READY"
	FileStatusPendingDelete FileStatus = "PENDING_DELETE"
	FileStatusDeleted       FileStatus = "DELETED"
)

type FileUsageType string

const (
	FileUsageTypeUserAvatar     FileUsageType = "USER_AVATAR"
	FileUsageTypePetAvatar      FileUsageType = "PET_AVATAR"
	FileUsageTypeLogAttachment  FileUsageType = "LOG_ATTACHMENT"
	FileUsageTypeHealthAttach   FileUsageType = "HEALTH_ATTACHMENT"
	FileUsageTypeChatAttachment FileUsageType = "CHAT_ATTACHMENT"
)

type FileObject struct {
	ID               uuid.UUID  `json:"id"`
	StorageBucket    string     `json:"storage_bucket"`
	StorageKey       string     `json:"storage_key"`
	Status           FileStatus `json:"status"`
	MimeType         string     `json:"mime_type"`
	SizeBytes        *int64     `json:"size_bytes"`
	OriginalFilename *string    `json:"original_filename"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	UploadExpiresAt  time.Time  `json:"upload_expires_at"`
	DeletedAt        *time.Time `json:"deleted_at"`
}

type ObjectInfo struct {
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type FileLink struct {
	ID        uuid.UUID     `json:"id"`
	FileID    uuid.UUID     `json:"file_id"`
	UsageType FileUsageType `json:"usage_type"`
	OwnerID   uuid.UUID     `json:"owner_id"`
	CreatedAt time.Time     `json:"created_at"`
}
