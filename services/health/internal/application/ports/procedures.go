package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type ListProceduresQuery struct {
	PetID               uuid.UUID
	Cursor              *TimeCursor
	Limit               int
	Q                   string
	Status              *string
	Bucket              string
	ProcedureTypeItemID *uuid.UUID
	DateFrom            *time.Time
	DateTo              *time.Time
	Sort                string
}

type ListProceduresResult struct {
	Items      []model.ProcedureListItem
	NextCursor *TimeCursor
}

type CreateProcedureInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	GeneratedFromID     *uuid.UUID
	Status              string
	ProcedureTypeItemID *uuid.UUID
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
	ProcedureTypeItemID *uuid.UUID
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

type UpdateGeneratedProcedurePlanInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	ProcedureTypeItemID *uuid.UUID
	Title               string
	Description         *string
	CatalogMedicationID *uuid.UUID
	ProductName         *string
	ScheduledAt         *time.Time
	UpdatedBy           uuid.UUID
}

type DeleteProcedureInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type ProcedureRepository interface {
	GetProcedure(ctx context.Context, petID, procedureID uuid.UUID) (*model.Procedure, error)
	GetGeneratedProcedure(ctx context.Context, petID, generatedFromID uuid.UUID) (*model.Procedure, error)
	ListProcedures(ctx context.Context, query ListProceduresQuery) (ListProceduresResult, error)
	CreateProcedure(ctx context.Context, input CreateProcedureInput) (*model.Procedure, AttachmentSync, error)
	UpdateProcedure(ctx context.Context, input UpdateProcedureInput) (*model.Procedure, AttachmentSync, error)
	UpdateGeneratedProcedurePlan(ctx context.Context, input UpdateGeneratedProcedurePlanInput) (*model.Procedure, error)
	DeleteProcedure(ctx context.Context, input DeleteProcedureInput) error
}
