package usecase

import (
	"health/internal/domain/model"
	"testing"
	"time"
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
