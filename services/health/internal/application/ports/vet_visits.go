package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type AttachmentSync struct {
	Add    []uuid.UUID
	Remove []uuid.UUID
}

type ListVetVisitsQuery struct {
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

type ListVetVisitsResult struct {
	Items      []model.VetVisitListItem
	NextCursor *TimeCursor
}

type CreateVetVisitInput struct {
	ID          uuid.UUID
	PetID       uuid.UUID
	Status      string
	VisitType   string
	Title       *string
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
	Title       *string
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

type VetVisitRepository interface {
	GetVetVisit(ctx context.Context, petID, visitID uuid.UUID, includeRelatedLogs bool) (*model.VetVisit, error)
	ListVetVisits(ctx context.Context, query ListVetVisitsQuery) (ListVetVisitsResult, error)
	CreateVetVisit(ctx context.Context, input CreateVetVisitInput) (*model.VetVisit, AttachmentSync, error)
	UpdateVetVisit(ctx context.Context, input UpdateVetVisitInput) (*model.VetVisit, AttachmentSync, error)
	DeleteVetVisit(ctx context.Context, input DeleteVetVisitInput) error
	LinkVetVisitLog(ctx context.Context, petID, visitID, logID, addedBy uuid.UUID) (*model.RelatedLog, error)
	UnlinkVetVisitLog(ctx context.Context, petID, visitID, logID uuid.UUID) error
}
