package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ACLPolicy struct {
	PetRead      bool
	PetWrite     bool
	LogRead      bool
	LogWrite     bool
	HealthRead   bool
	HealthWrite  bool
	MembersRead  bool
	MembersWrite bool
}

type ACLRole struct {
	ID              uuid.UUID
	Kind            string
	PetID           *uuid.UUID
	Code            *string
	Title           string
	Policy          ACLPolicy
	CreatedByUserID *uuid.UUID
}

type ACLMembership struct {
	PetID          uuid.UUID
	MemberID       uuid.UUID
	Status         string
	IsPrimaryOwner bool
	Role           ACLRole
	Policy         ACLPolicy
}

type ACLTransferOwnershipResult struct {
	PreviousOwnerMemberID uuid.UUID
	PreviousOwnerUserID   uuid.UUID
	CurrentOwnerMemberID  uuid.UUID
	CurrentOwnerUserID    uuid.UUID
}

type UploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

type ACLClient interface {
	Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error)
	ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]ACLMembership, error)
	CreateOwnerMembership(ctx context.Context, petID, userID uuid.UUID) (uuid.UUID, error)
	TransferOwnership(ctx context.Context, petID, requesterUserID, targetMemberID uuid.UUID) (ACLTransferOwnershipResult, error)
}

type FileClient interface {
	InitUpload(ctx context.Context, mimeType string, expectedSize int64, originalFilename string) (uuid.UUID, UploadInfo, error)
	ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) error
	GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error)
	BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error)
	LinkPetAvatar(ctx context.Context, fileID, petID uuid.UUID) error
	UnlinkPetAvatar(ctx context.Context, fileID, petID uuid.UUID) error
	DeleteFileIfUnlinked(ctx context.Context, fileID uuid.UUID) error
}
