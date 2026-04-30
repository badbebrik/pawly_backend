package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type ListDueScheduledItemOccurrencesParams struct {
	Before      time.Time
	Limit       int
	DispatchKey string
}

type CreateScheduledItemDispatchParams struct {
	ID                        uuid.UUID
	ScheduledItemOccurrenceID uuid.UUID
	DispatchKey               string
}

type ScheduledDispatchRepository interface {
	ListDueScheduledItemOccurrences(ctx context.Context, params ListDueScheduledItemOccurrencesParams) ([]model.ScheduledItemOccurrenceListItem, error)
	CreateScheduledItemDispatch(ctx context.Context, params CreateScheduledItemDispatchParams) error
}

type PetUserLister interface {
	ListPetUserIDs(ctx context.Context, petID uuid.UUID) ([]uuid.UUID, error)
}

type PushPublisher interface {
	PublishScheduledOccurrenceDue(ctx context.Context, job model.ScheduledOccurrencePushJob) error
}
