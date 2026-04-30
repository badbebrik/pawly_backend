package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type ListMedicalRecordsQuery struct {
	PetID            uuid.UUID
	Cursor           *TimeCursor
	Limit            int
	Q                string
	Status           *string
	Bucket           string
	RecordTypeItemID *uuid.UUID
	Sort             string
}

type ListMedicalRecordsResult struct {
	Items      []model.MedicalRecordListItem
	NextCursor *TimeCursor
}

type CreateMedicalRecordInput struct {
	ID               uuid.UUID
	PetID            uuid.UUID
	RecordTypeItemID *uuid.UUID
	Status           string
	Title            string
	Description      *string
	StartedAt        *time.Time
	ResolvedAt       *time.Time
	CreatedBy        uuid.UUID
	UpdatedBy        uuid.UUID
	Attachments      []AttachmentInput
}

type UpdateMedicalRecordInput struct {
	ID               uuid.UUID
	PetID            uuid.UUID
	RowVersion       int
	RecordTypeItemID *uuid.UUID
	Status           string
	Title            string
	Description      *string
	StartedAt        *time.Time
	ResolvedAt       *time.Time
	UpdatedBy        uuid.UUID
	Attachments      []AttachmentInput
}

type DeleteMedicalRecordInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type MedicalRecordRepository interface {
	GetMedicalRecord(ctx context.Context, petID, recordID uuid.UUID) (*model.MedicalRecord, error)
	ListMedicalRecords(ctx context.Context, query ListMedicalRecordsQuery) (ListMedicalRecordsResult, error)
	CreateMedicalRecord(ctx context.Context, input CreateMedicalRecordInput) (*model.MedicalRecord, AttachmentSync, error)
	UpdateMedicalRecord(ctx context.Context, input UpdateMedicalRecordInput) (*model.MedicalRecord, AttachmentSync, error)
	DeleteMedicalRecord(ctx context.Context, input DeleteMedicalRecordInput) error
}
