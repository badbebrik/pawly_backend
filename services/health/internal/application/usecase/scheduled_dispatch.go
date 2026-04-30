package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

const scheduledOccurrenceDispatchKey = "scheduled_occurrence_due"

type ScheduledDispatcherDependencies struct {
	Repository    ports.ScheduledDispatchRepository
	PetUserLister ports.PetUserLister
	PushPublisher ports.PushPublisher
}

type DispatchFailure struct {
	Operation    string
	PetID        uuid.UUID
	OccurrenceID uuid.UUID
	Err          error
}

type DispatchDueScheduledOccurrencesResult struct {
	Scanned   int
	Published int
	Skipped   int
	Failed    int
	Failures  []DispatchFailure
}

type ScheduledDispatcher struct {
	repository    ports.ScheduledDispatchRepository
	petUserLister ports.PetUserLister
	pushPublisher ports.PushPublisher
}

func NewScheduledDispatcher(deps ScheduledDispatcherDependencies) *ScheduledDispatcher {
	return &ScheduledDispatcher{
		repository:    deps.Repository,
		petUserLister: deps.PetUserLister,
		pushPublisher: deps.PushPublisher,
	}
}

func (u *ScheduledDispatcher) DispatchDueScheduledOccurrences(ctx context.Context, batchSize int) (DispatchDueScheduledOccurrencesResult, error) {
	items, err := u.repository.ListDueScheduledItemOccurrences(ctx, ports.ListDueScheduledItemOccurrencesParams{
		Before:      time.Now().UTC(),
		Limit:       batchSize,
		DispatchKey: scheduledOccurrenceDispatchKey,
	})
	if err != nil {
		return DispatchDueScheduledOccurrencesResult{}, err
	}

	result := DispatchDueScheduledOccurrencesResult{Scanned: len(items)}
	for i := range items {
		item := items[i]

		userIDs, err := u.petUserLister.ListPetUserIDs(ctx, item.PetID)
		if err != nil {
			result.Failed++
			result.Failures = append(result.Failures, DispatchFailure{
				Operation:    "list_pet_user_ids",
				PetID:        item.PetID,
				OccurrenceID: item.ID,
				Err:          err,
			})
			continue
		}
		if len(userIDs) == 0 {
			result.Skipped++
			continue
		}

		err = u.repository.CreateScheduledItemDispatch(ctx, ports.CreateScheduledItemDispatchParams{
			ID:                        uuid.New(),
			ScheduledItemOccurrenceID: item.ID,
			DispatchKey:               scheduledOccurrenceDispatchKey,
		})
		if err != nil {
			if err == ports.ErrConflict {
				result.Skipped++
				continue
			}
			result.Failed++
			result.Failures = append(result.Failures, DispatchFailure{
				Operation:    "create_dispatch",
				PetID:        item.PetID,
				OccurrenceID: item.ID,
				Err:          err,
			})
			continue
		}

		job := model.ScheduledOccurrencePushJob{
			Event:           "SCHEDULED_OCCURRENCE_DUE",
			PetID:           item.PetID.String(),
			OccurrenceID:    item.ID.String(),
			ScheduledItemID: item.ScheduledItemID.String(),
			UserIDs:         uuidStrings(userIDs),
			SourceType:      item.Rule.SourceType,
			Title:           item.Rule.Title,
			ScheduledFor:    item.ScheduledFor.UTC().Format(time.RFC3339),
		}
		if item.Rule.Note != nil {
			job.Note = *item.Rule.Note
		}

		if err := u.pushPublisher.PublishScheduledOccurrenceDue(ctx, job); err != nil {
			result.Failed++
			result.Failures = append(result.Failures, DispatchFailure{
				Operation:    "publish_push_job",
				PetID:        item.PetID,
				OccurrenceID: item.ID,
				Err:          err,
			})
			continue
		}

		result.Published++
	}

	return result, nil
}

func uuidStrings(items []uuid.UUID) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		out = append(out, item.String())
	}
	return out
}
