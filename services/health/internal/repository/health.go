package repository

import (
	"context"
	"health/internal/model"
	"time"

	"github.com/google/uuid"
)

type TimeCursor struct {
	SortAt time.Time
	ID     uuid.UUID
}

type AttachmentInput struct {
	FileID   uuid.UUID
	FileName *string
	FileType string
}

type AttachmentSync struct {
	Add    []uuid.UUID
	Remove []uuid.UUID
}

type ListVetVisitsInput struct {
	PetID    uuid.UUID
	Cursor   *TimeCursor
	Limit    int
	Status   *string
	Bucket   string
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     string
}

type ListVetVisitsOutput struct {
	Items      []model.VetVisitListItem
	NextCursor *TimeCursor
}

type CreateVetVisitInput struct {
	ID          uuid.UUID
	PetID       uuid.UUID
	Status      string
	VisitType   string
	ScheduledAt *time.Time
	CompletedAt *time.Time
	ReasonText  *string
	ResultText  *string
	ClinicName  *string
	VetName     *string
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
	Attachments []AttachmentInput
}

type UpdateVetVisitInput struct {
	ID          uuid.UUID
	PetID       uuid.UUID
	RowVersion  int
	Status      string
	VisitType   string
	ScheduledAt *time.Time
	CompletedAt *time.Time
	ReasonText  *string
	ResultText  *string
	ClinicName  *string
	VetName     *string
	UpdatedBy   uuid.UUID
	Attachments []AttachmentInput
}

type DeleteVetVisitInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type CreateVaccinationInput struct {
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
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
	Attachments         []AttachmentInput
}

type UpdateVaccinationInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	RowVersion          int
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
	UpdatedBy           uuid.UUID
	Attachments         []AttachmentInput
}

type DeleteVaccinationInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type ListVaccinationsInput struct {
	PetID    uuid.UUID
	Cursor   *TimeCursor
	Limit    int
	Status   *string
	Bucket   string
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     string
}

type ListVaccinationsOutput struct {
	Items      []model.VaccinationListItem
	NextCursor *TimeCursor
}

type CreateProcedureInput struct {
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
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
	Attachments         []AttachmentInput
}

type UpdateProcedureInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	RowVersion          int
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
	UpdatedBy           uuid.UUID
	Attachments         []AttachmentInput
}

type DeleteProcedureInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type ListProceduresInput struct {
	PetID         uuid.UUID
	Cursor        *TimeCursor
	Limit         int
	Status        *string
	Bucket        string
	ProcedureType *string
	DateFrom      *time.Time
	DateTo        *time.Time
	Sort          string
}

type ListProceduresOutput struct {
	Items      []model.ProcedureListItem
	NextCursor *TimeCursor
}

type CreateMedicalRecordInput struct {
	ID          uuid.UUID
	PetID       uuid.UUID
	RecordType  string
	Status      string
	Title       string
	Description *string
	StartedAt   *time.Time
	ResolvedAt  *time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
	Attachments []AttachmentInput
}

type UpdateMedicalRecordInput struct {
	ID          uuid.UUID
	PetID       uuid.UUID
	RowVersion  int
	RecordType  string
	Status      string
	Title       string
	Description *string
	StartedAt   *time.Time
	ResolvedAt  *time.Time
	UpdatedBy   uuid.UUID
	Attachments []AttachmentInput
}

type DeleteMedicalRecordInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type ListMedicalRecordsInput struct {
	PetID      uuid.UUID
	Cursor     *TimeCursor
	Limit      int
	Status     *string
	Bucket     string
	RecordType *string
	Sort       string
}

type ListMedicalRecordsOutput struct {
	Items      []model.MedicalRecordListItem
	NextCursor *TimeCursor
}

type ListPetDocumentsInput struct {
	PetID      uuid.UUID
	Cursor     *TimeCursor
	Limit      int
	EntityType *string
	FileType   *string
}

type ListPetDocumentsOutput struct {
	Items      []model.PetDocument
	NextCursor *TimeCursor
}

type ListScheduledItemsInput struct {
	PetID       uuid.UUID
	Cursor      *TimeCursor
	Limit       int
	SourceType  *string
	DateFrom    *time.Time
	DateTo      *time.Time
	IncludePast bool
}

type ListScheduledItemsOutput struct {
	Items      []model.ScheduledItemListItem
	NextCursor *TimeCursor
}

type CreateScheduledItemInput struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	SourceType         string
	SourceID           *uuid.UUID
	Title              string
	Note               *string
	StartsAt           time.Time
	RecurrenceRule     *string
	RecurrenceInterval *int
	RecurrenceUntil    *time.Time
	CreatedBy          uuid.UUID
	UpdatedBy          uuid.UUID
}

type UpdateScheduledItemInput struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	RowVersion         int
	Title              string
	Note               *string
	StartsAt           time.Time
	RecurrenceRule     *string
	RecurrenceInterval *int
	RecurrenceUntil    *time.Time
	UpdatedBy          uuid.UUID
}

type DeleteScheduledItemInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type UpsertHealthScheduledItemInput struct {
	PetID           uuid.UUID
	SourceType      string
	SourceID        uuid.UUID
	Title           string
	Note            *string
	StartsAt        time.Time
	CreatedByUserID uuid.UUID
	UpdatedByUserID uuid.UUID
}

type DeleteHealthScheduledItemInput struct {
	PetID           uuid.UUID
	SourceType      string
	SourceID        uuid.UUID
	DeletedByUserID uuid.UUID
}

type CreateScheduledItemDispatchInput struct {
	ID                        uuid.UUID
	ScheduledItemOccurrenceID uuid.UUID
	DispatchKey               string
}

type ListScheduledItemOccurrencesInput struct {
	PetID      uuid.UUID
	Cursor     *TimeCursor
	Limit      int
	DateFrom   *time.Time
	DateTo     *time.Time
	SourceType *string
}

type ListScheduledItemOccurrencesOutput struct {
	Items      []model.ScheduledItemOccurrenceListItem
	NextCursor *TimeCursor
}

type CreateScheduledItemOccurrenceInput struct {
	ID              uuid.UUID
	ScheduledItemID uuid.UUID
	PetID           uuid.UUID
	ScheduledFor    time.Time
}

type DeleteScheduledItemOccurrencesFromInput struct {
	ScheduledItemID uuid.UUID
	From            time.Time
}

type HealthRepository interface {
	GetVetVisit(ctx context.Context, petID, visitID uuid.UUID, includeRelatedLogs bool) (*model.VetVisit, error)
	ListVetVisits(ctx context.Context, in ListVetVisitsInput) (ListVetVisitsOutput, error)
	CreateVetVisit(ctx context.Context, in CreateVetVisitInput) (*model.VetVisit, AttachmentSync, error)
	UpdateVetVisit(ctx context.Context, in UpdateVetVisitInput) (*model.VetVisit, AttachmentSync, error)
	DeleteVetVisit(ctx context.Context, in DeleteVetVisitInput) error
	LinkVetVisitLog(ctx context.Context, petID, visitID, logID, addedBy uuid.UUID) (*model.RelatedLog, error)
	UnlinkVetVisitLog(ctx context.Context, petID, visitID, logID uuid.UUID) error

	GetVaccination(ctx context.Context, petID, vaccinationID uuid.UUID) (*model.Vaccination, error)
	ListVaccinations(ctx context.Context, in ListVaccinationsInput) (ListVaccinationsOutput, error)
	CreateVaccination(ctx context.Context, in CreateVaccinationInput) (*model.Vaccination, AttachmentSync, error)
	UpdateVaccination(ctx context.Context, in UpdateVaccinationInput) (*model.Vaccination, AttachmentSync, error)
	DeleteVaccination(ctx context.Context, in DeleteVaccinationInput) error

	GetProcedure(ctx context.Context, petID, procedureID uuid.UUID) (*model.Procedure, error)
	ListProcedures(ctx context.Context, in ListProceduresInput) (ListProceduresOutput, error)
	CreateProcedure(ctx context.Context, in CreateProcedureInput) (*model.Procedure, AttachmentSync, error)
	UpdateProcedure(ctx context.Context, in UpdateProcedureInput) (*model.Procedure, AttachmentSync, error)
	DeleteProcedure(ctx context.Context, in DeleteProcedureInput) error

	GetMedicalRecord(ctx context.Context, petID, recordID uuid.UUID) (*model.MedicalRecord, error)
	ListMedicalRecords(ctx context.Context, in ListMedicalRecordsInput) (ListMedicalRecordsOutput, error)
	CreateMedicalRecord(ctx context.Context, in CreateMedicalRecordInput) (*model.MedicalRecord, AttachmentSync, error)
	UpdateMedicalRecord(ctx context.Context, in UpdateMedicalRecordInput) (*model.MedicalRecord, AttachmentSync, error)
	DeleteMedicalRecord(ctx context.Context, in DeleteMedicalRecordInput) error
	ListPetDocuments(ctx context.Context, in ListPetDocumentsInput) (ListPetDocumentsOutput, error)

	GetScheduledItem(ctx context.Context, petID, itemID uuid.UUID) (*model.ScheduledItem, error)
	ListScheduledItems(ctx context.Context, in ListScheduledItemsInput) (ListScheduledItemsOutput, error)
	CreateScheduledItem(ctx context.Context, in CreateScheduledItemInput) (*model.ScheduledItem, error)
	UpdateScheduledItem(ctx context.Context, in UpdateScheduledItemInput) (*model.ScheduledItem, error)
	DeleteScheduledItem(ctx context.Context, in DeleteScheduledItemInput) error
	UpsertHealthScheduledItem(ctx context.Context, in UpsertHealthScheduledItemInput) (*model.ScheduledItem, error)
	DeleteHealthScheduledItem(ctx context.Context, in DeleteHealthScheduledItemInput) error
	GetScheduledItemOccurrence(ctx context.Context, petID, occurrenceID uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error)
	ListScheduledItemOccurrences(ctx context.Context, in ListScheduledItemOccurrencesInput) (ListScheduledItemOccurrencesOutput, error)
	CreateScheduledItemOccurrence(ctx context.Context, in CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error)
	DeleteScheduledItemOccurrencesFrom(ctx context.Context, in DeleteScheduledItemOccurrencesFromInput) error
	CreateScheduledItemDispatch(ctx context.Context, in CreateScheduledItemDispatchInput) error

	ListCalendarDayItems(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.CalendarDayItem, error)
	ListCalendarDayScheduledOccurrences(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error)
	ListCalendarDayScheduledOccurrencesForPets(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error)
}

type Repository interface {
	LogRepository
	HealthRepository
}
