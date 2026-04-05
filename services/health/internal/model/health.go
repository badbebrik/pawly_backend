package model

import (
	"time"

	"github.com/google/uuid"
)

type HealthAttachment struct {
	ID            uuid.UUID
	EntityType    string
	EntityID      uuid.UUID
	FileID        uuid.UUID
	FileName      *string
	FileType      string
	AddedByUserID uuid.UUID
	AddedAt       time.Time
	DownloadURL   *string
	PreviewURL    *string
}

type PetDocument struct {
	ID            uuid.UUID
	PetID         uuid.UUID
	EntityType    string
	EntityID      uuid.UUID
	FileID        uuid.UUID
	FileName      *string
	FileType      string
	AddedByUserID uuid.UUID
	AddedAt       time.Time
	DownloadURL   *string
	PreviewURL    *string
}

type RelatedLog struct {
	ID                 uuid.UUID
	OccurredAt         time.Time
	LogTypeName        *string
	DescriptionPreview *string
	Source             string
}

type HealthPermissions struct {
	HealthRead  bool
	HealthWrite bool
	LogRead     bool
}

type HealthBootstrap struct {
	Permissions HealthPermissions
	Enums       HealthEnums
}

type HealthEnums struct {
	VetVisitStatuses      []string
	VetVisitTypes         []string
	VaccinationStatuses   []string
	ProcedureStatuses     []string
	ProcedureTypes        []string
	MedicalRecordTypes    []string
	MedicalRecordStatuses []string
}

type VetVisit struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	Status          string
	VisitType       string
	ScheduledAt     *time.Time
	CompletedAt     *time.Time
	ReasonText      *string
	ResultText      *string
	ClinicName      *string
	VetName         *string
	RelatedLogs     []RelatedLog
	Attachments     []HealthAttachment
	RowVersion      int
	CreatedAt       time.Time
	CreatedByUserID uuid.UUID
	UpdatedAt       time.Time
	UpdatedByUserID uuid.UUID
}

type VetVisitListItem struct {
	ID               uuid.UUID
	PetID            uuid.UUID
	Status           string
	VisitType        string
	ScheduledAt      *time.Time
	CompletedAt      *time.Time
	ReasonText       *string
	ResultText       *string
	ClinicName       *string
	VetName          *string
	RelatedLogsCount int
	AttachmentsCount int
	RowVersion       int
	CreatedAt        time.Time
	CreatedByUserID  uuid.UUID
	UpdatedAt        time.Time
	UpdatedByUserID  uuid.UUID
}

type Vaccination struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	ScheduledAt         *time.Time
	AdministeredAt      *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	ClinicName          *string
	VetName             *string
	Notes               *string
	Attachments         []HealthAttachment
	RowVersion          int
	CreatedAt           time.Time
	CreatedByUserID     uuid.UUID
	UpdatedAt           time.Time
	UpdatedByUserID     uuid.UUID
}

type VaccinationListItem struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	ScheduledAt         *time.Time
	AdministeredAt      *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	ClinicName          *string
	VetName             *string
	NotesPreview        *string
	AttachmentsCount    int
	RowVersion          int
	CreatedAt           time.Time
	CreatedByUserID     uuid.UUID
	UpdatedAt           time.Time
	UpdatedByUserID     uuid.UUID
}

type Procedure struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	Status              string
	ProcedureType       string
	Title               string
	Description         *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	PerformedAt         *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	Notes               *string
	Attachments         []HealthAttachment
	RowVersion          int
	CreatedAt           time.Time
	CreatedByUserID     uuid.UUID
	UpdatedAt           time.Time
	UpdatedByUserID     uuid.UUID
}

type ProcedureListItem struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	Status              string
	ProcedureType       string
	Title               string
	DescriptionPreview  *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	PerformedAt         *time.Time
	NextDueAt           *time.Time
	VetVisitID          *uuid.UUID
	NotesPreview        *string
	AttachmentsCount    int
	RowVersion          int
	CreatedAt           time.Time
	CreatedByUserID     uuid.UUID
	UpdatedAt           time.Time
	UpdatedByUserID     uuid.UUID
}

type MedicalRecord struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RecordType      string
	Status          string
	Title           string
	Description     *string
	StartedAt       *time.Time
	ResolvedAt      *time.Time
	Attachments     []HealthAttachment
	RowVersion      int
	CreatedAt       time.Time
	CreatedByUserID uuid.UUID
	UpdatedAt       time.Time
	UpdatedByUserID uuid.UUID
}

type MedicalRecordListItem struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	RecordType         string
	Status             string
	Title              string
	DescriptionPreview *string
	StartedAt          *time.Time
	ResolvedAt         *time.Time
	AttachmentsCount   int
	RowVersion         int
	CreatedAt          time.Time
	CreatedByUserID    uuid.UUID
	UpdatedAt          time.Time
	UpdatedByUserID    uuid.UUID
}

type CalendarDayItem struct {
	ItemType      string
	EntityID      uuid.UUID
	Title         string
	Subtitle      *string
	ScheduledFor  time.Time
	Status        string
	VisitID       *uuid.UUID
	VaccinationID *uuid.UUID
	ProcedureID   *uuid.UUID
}

type HealthFile struct {
	ID       uuid.UUID
	MimeType string
	FileName *string
}
