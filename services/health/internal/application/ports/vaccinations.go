package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type ListVaccinationsQuery struct {
	PetID    uuid.UUID
	Cursor   *TimeCursor
	Limit    int
	Q        string
	Status   *string
	Bucket   string
	DateFrom *time.Time
	DateTo   *time.Time
	Sort     string
}

type ListVaccinationsResult struct {
	Items      []model.VaccinationListItem
	NextCursor *TimeCursor
}

type CreateVaccinationInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	GeneratedFromID     *uuid.UUID
	Status              string
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	TargetItemIDs       []uuid.UUID
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
	TargetItemIDs       []uuid.UUID
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

type UpdateGeneratedVaccinationPlanInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	VaccineName         string
	CatalogMedicationID *uuid.UUID
	TargetItemIDs       []uuid.UUID
	ScheduledAt         *time.Time
	UpdatedBy           uuid.UUID
}

type DeleteVaccinationInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type VaccinationRepository interface {
	GetVaccination(ctx context.Context, petID, vaccinationID uuid.UUID) (*model.Vaccination, error)
	GetGeneratedVaccination(ctx context.Context, petID, generatedFromID uuid.UUID) (*model.Vaccination, error)
	ListVaccinations(ctx context.Context, query ListVaccinationsQuery) (ListVaccinationsResult, error)
	CreateVaccination(ctx context.Context, input CreateVaccinationInput) (*model.Vaccination, AttachmentSync, error)
	UpdateVaccination(ctx context.Context, input UpdateVaccinationInput) (*model.Vaccination, AttachmentSync, error)
	UpdateGeneratedVaccinationPlan(ctx context.Context, input UpdateGeneratedVaccinationPlanInput) (*model.Vaccination, error)
	DeleteVaccination(ctx context.Context, input DeleteVaccinationInput) error
}
