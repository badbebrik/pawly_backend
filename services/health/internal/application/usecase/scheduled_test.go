package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildScheduledOccurrencesSkipsPastRecurringOccurrences(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	rule := model.RecurrenceRuleDaily
	interval := 1
	item := &model.ScheduledItem{
		StartsAt:           now.AddDate(0, 0, -10),
		RecurrenceRule:     &rule,
		RecurrenceInterval: &interval,
	}

	items := buildScheduledOccurrences(item, now)
	if len(items) == 0 {
		t.Fatal("expected occurrences")
	}
	for _, scheduledFor := range items {
		if scheduledFor.Before(now) {
			t.Fatalf("expected no past occurrence, got %s before %s", scheduledFor, now)
		}
	}
}

func TestBuildScheduledOccurrencesSkipsPastOneShot(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	item := &model.ScheduledItem{StartsAt: now.Add(-time.Hour)}

	items := buildScheduledOccurrences(item, now)
	if len(items) != 0 {
		t.Fatalf("expected no occurrence for past one-shot, got %d", len(items))
	}
}

func TestBuildScheduledOccurrencesKeepsFutureOneShot(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	startsAt := now.Add(time.Hour)
	item := &model.ScheduledItem{StartsAt: startsAt}

	items := buildScheduledOccurrences(item, now)
	if len(items) != 1 {
		t.Fatalf("expected one future occurrence, got %d", len(items))
	}
	if !items[0].Equal(startsAt) {
		t.Fatalf("unexpected occurrence time: got %s want %s", items[0], startsAt)
	}
}

func TestScheduledCreateManualItemCreatesOccurrence(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	startsAt := time.Now().UTC().Add(time.Hour)
	var createdInput ports.CreateScheduledItemInput
	var occurrenceInput ports.CreateScheduledItemOccurrenceInput
	deletedOccurrences := false

	repo := &stubScheduledRepo{
		createScheduledItemFn: func(_ context.Context, input ports.CreateScheduledItemInput) (*model.ScheduledItem, error) {
			createdInput = input
			return &model.ScheduledItem{
				ID:                  input.ID,
				PetID:               input.PetID,
				SourceType:          input.SourceType,
				Title:               input.Title,
				StartsAt:            input.StartsAt,
				PushEnabled:         input.PushEnabled,
				RemindOffsetMinutes: input.RemindOffsetMinutes,
				RowVersion:          1,
			}, nil
		},
		deleteScheduledItemOccurrencesFromFn: func(context.Context, ports.DeleteScheduledItemOccurrencesFromInput) error {
			deletedOccurrences = true
			return nil
		},
		createScheduledItemOccurrenceFn: func(_ context.Context, input ports.CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error) {
			occurrenceInput = input
			return &model.ScheduledItemOccurrence{ID: input.ID, PetID: input.PetID, ScheduledItemID: input.ScheduledItemID, ScheduledFor: input.ScheduledFor}, nil
		},
	}
	uc := NewScheduled(repo, &stubHealthAccess{})

	out, err := uc.CreateScheduledItem(context.Background(), CreateScheduledItemParams{
		UserID:     userID,
		PetID:      petID,
		SourceType: model.ScheduledItemSourceTypeManual,
		Title:      " Walk ",
		StartsAt:   startsAt,
	})
	if err != nil {
		t.Fatalf("CreateScheduledItem returned error: %v", err)
	}
	if out.ID == uuid.Nil || createdInput.ID != out.ID || createdInput.Title != "Walk" {
		t.Fatalf("unexpected created item: input=%+v out=%+v", createdInput, out)
	}
	if !createdInput.PushEnabled || createdInput.RemindOffsetMinutes == nil || *createdInput.RemindOffsetMinutes != 0 {
		t.Fatalf("unexpected default reminder settings: %+v", createdInput)
	}
	if !deletedOccurrences {
		t.Fatal("expected old occurrences to be deleted before regeneration")
	}
	if occurrenceInput.ScheduledItemID != out.ID || !occurrenceInput.ScheduledFor.Equal(startsAt.UTC()) {
		t.Fatalf("unexpected occurrence input: %+v", occurrenceInput)
	}
}

func TestScheduledCreateRejectsInvalidSourceAndReminder(t *testing.T) {
	uc := NewScheduled(&stubScheduledRepo{}, &stubHealthAccess{})
	userID := uuid.New()
	petID := uuid.New()
	startsAt := time.Now().UTC().Add(time.Hour)

	_, err := uc.CreateScheduledItem(context.Background(), CreateScheduledItemParams{
		UserID:     userID,
		PetID:      petID,
		SourceType: model.ScheduledItemSourceTypeLogType,
		Title:      "Log",
		StartsAt:   startsAt,
	})
	expectHealthErr(t, err, ErrInvalidInput)

	_, err = uc.CreateScheduledItem(context.Background(), CreateScheduledItemParams{
		UserID:              userID,
		PetID:               petID,
		SourceType:          model.ScheduledItemSourceTypeManual,
		Title:               "Walk",
		StartsAt:            startsAt,
		RemindOffsetMinutes: intPtr(-1),
	})
	expectHealthErr(t, err, ErrInvalidInput)
}

func TestScheduledListUsesAllowedSourceTypes(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	var captured ports.ListScheduledItemsQuery
	repo := &stubScheduledRepo{
		listScheduledItemsFn: func(_ context.Context, query ports.ListScheduledItemsQuery) (ports.ListScheduledItemsResult, error) {
			captured = query
			return ports.ListScheduledItemsResult{}, nil
		},
	}
	acl := &stubHealthAccess{allowed: map[string]bool{
		ActionPetRead:    true,
		ActionLogRead:    false,
		ActionHealthRead: true,
	}}
	uc := NewScheduled(repo, acl)

	_, err := uc.ListScheduledItems(context.Background(), ListScheduledItemsParams{
		UserID: userID,
		PetID:  petID,
	})
	if err != nil {
		t.Fatalf("ListScheduledItems returned error: %v", err)
	}

	got := map[string]bool{}
	for _, sourceType := range captured.SourceTypes {
		got[sourceType] = true
	}
	if !got[model.ScheduledItemSourceTypeManual] || !got[model.ScheduledItemSourceTypePetEvent] || !got[model.ScheduledItemSourceTypeVetVisit] || !got[model.ScheduledItemSourceTypeVaccination] || !got[model.ScheduledItemSourceTypeProcedure] {
		t.Fatalf("missing expected source types: %+v", captured.SourceTypes)
	}
	if got[model.ScheduledItemSourceTypeLogType] {
		t.Fatalf("did not expect log type source without log_read: %+v", captured.SourceTypes)
	}
}

func TestScheduledUpdateRejectsSystemSource(t *testing.T) {
	itemID := uuid.New()
	uc := NewScheduled(&stubScheduledRepo{
		getScheduledItemFn: func(context.Context, uuid.UUID, uuid.UUID) (*model.ScheduledItem, error) {
			return &model.ScheduledItem{ID: itemID, SourceType: model.ScheduledItemSourceTypeVetVisit}, nil
		},
	}, &stubHealthAccess{})

	_, err := uc.UpdateScheduledItem(context.Background(), UpdateScheduledItemParams{
		UserID:     uuid.New(),
		PetID:      uuid.New(),
		ItemID:     itemID,
		RowVersion: 1,
		Title:      "Visit",
		StartsAt:   time.Now().UTC().Add(time.Hour),
	})
	expectHealthErr(t, err, ErrForbidden)
}

func TestScheduledGenerateOccurrencesHorizonCountsFailures(t *testing.T) {
	rule := model.RecurrenceRuleDaily
	interval := 1
	repo := &stubScheduledRepo{
		listRecurringScheduledItemsForHorizonFn: func(context.Context, ports.ListRecurringScheduledItemsForHorizonParams) ([]model.ScheduledItem, error) {
			return []model.ScheduledItem{
				{ID: uuid.New(), PetID: uuid.New(), StartsAt: time.Now().UTC(), RecurrenceRule: &rule, RecurrenceInterval: &interval},
				{ID: uuid.New(), PetID: uuid.New(), StartsAt: time.Now().UTC(), RecurrenceRule: &rule, RecurrenceInterval: &interval},
			}, nil
		},
		createScheduledItemOccurrenceFn: func(context.Context, ports.CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error) {
			return nil, ports.ErrInvalidInput
		},
	}
	uc := NewScheduled(repo, &stubHealthAccess{})

	out, err := uc.GenerateOccurrencesHorizon(context.Background(), 0)
	if err != nil {
		t.Fatalf("GenerateOccurrencesHorizon returned error: %v", err)
	}
	if out.Scanned != 2 || out.Failed != 2 {
		t.Fatalf("unexpected horizon result: %+v", out)
	}
}
