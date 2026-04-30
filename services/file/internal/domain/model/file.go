package model

import (
	"time"

	"github.com/google/uuid"
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
	ID               uuid.UUID
	StorageBucket    string
	StorageKey       string
	Status           FileStatus
	MimeType         string
	SizeBytes        *int64
	OriginalFilename *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	UploadExpiresAt  time.Time
	DeletedAt        *time.Time
}

type ObjectInfo struct {
	Size        int64
	ContentType string
}

type FileLink struct {
	ID        uuid.UUID
	FileID    uuid.UUID
	UsageType FileUsageType
	OwnerID   uuid.UUID
	CreatedAt time.Time
}
