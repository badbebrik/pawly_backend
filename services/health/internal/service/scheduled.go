package service

import (
	"context"
	"health/internal/model"
	"health/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

const scheduledOccurrencesHorizonDays = 90

type ListScheduledItemsParams struct {
	UserID      uuid.UUID
	PetID       uuid.UUID
	Cursor      *repository.TimeCursor
	Limit       int
	SourceType  string
	DateFrom    *time.Time
	DateTo      *time.Time
	IncludePast bool
}

type CreateScheduledItemParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	SourceType          string
	SourceID            *uuid.UUID
	Title               string
	Note                *string
	StartsAt            time.Time
	PushEnabled         *bool
	RemindOffsetMinutes *int
	RecurrenceRule      *string
	RecurrenceInterval  *int
	RecurrenceUntil     *time.Time
}

type UpdateScheduledItemParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	ItemID              uuid.UUID
	RowVersion          int
	Title               string
	Note                *string
	StartsAt            time.Time
	PushEnabled         *bool
	RemindOffsetMinutes *int
	RecurrenceRule      *string
	RecurrenceInterval  *int
	RecurrenceUntil     *time.Time
}

type UpdateScheduledItemReminderSettingsParams struct {
	UserID              uuid.UUID
	PetID               uuid.UUID
	ItemID              uuid.UUID
	RowVersion          int
	PushEnabled         bool
	RemindOffsetMinutes *int
}

type DeleteScheduledItemParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	ItemID     uuid.UUID
	RowVersion int
}

type ListScheduledItemOccurrencesParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	Cursor     *repository.TimeCursor
	Limit      int
	SourceType string
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (s *Service) GetScheduledItem(ctx context.Context, userID, petID, itemID uuid.UUID) (*model.ScheduledItem, error) {
	if userID == uuid.Nil || petID == uuid.Nil || itemID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, petID, userID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetScheduledItem(ctx, petID, itemID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) ListScheduledItems(ctx context.Context, p ListScheduledItemsParams) (repository.ListScheduledItemsOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListScheduledItemsOutput{}, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, p.PetID, p.UserID); err != nil {
		return repository.ListScheduledItemsOutput{}, err
	}
	sourceType := normalizeEnumPtr(p.SourceType, []string{
		model.ScheduledItemSourceTypeManual,
		model.ScheduledItemSourceTypeLogType,
		model.ScheduledItemSourceTypePetEvent,
		model.ScheduledItemSourceTypeVetVisit,
		model.ScheduledItemSourceTypeVaccination,
		model.ScheduledItemSourceTypeProcedure,
	})
	if strings.TrimSpace(p.SourceType) != "" && sourceType == nil {
		return repository.ListScheduledItemsOutput{}, ErrInvalidInput
	}
	out, err := s.repo.ListScheduledItems(ctx, repository.ListScheduledItemsInput{
		PetID:       p.PetID,
		Cursor:      p.Cursor,
		Limit:       p.Limit,
		SourceType:  sourceType,
		DateFrom:    p.DateFrom,
		DateTo:      p.DateTo,
		IncludePast: p.IncludePast,
	})
	if err != nil {
		return repository.ListScheduledItemsOutput{}, mapRepoErr(err)
	}
	return out, nil
}

func (s *Service) CreateScheduledItem(ctx context.Context, p CreateScheduledItemParams) (*model.ScheduledItem, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.StartsAt.IsZero() {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	sourceType, err := validateScheduledItemWritableSourceType(p.SourceType)
	if err != nil {
		return nil, err
	}
	sourceID, err := validateScheduledItemSourceID(sourceType, p.SourceID)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	note := trimOrNil(p.Note)
	recurrenceRule, recurrenceInterval, recurrenceUntil, err := validateScheduledRecurrence(p.RecurrenceRule, p.RecurrenceInterval, p.RecurrenceUntil)
	if err != nil {
		return nil, err
	}
	pushEnabled, remindOffsetMinutes, err := validateScheduledReminderSettings(p.PushEnabled, p.RemindOffsetMinutes)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.CreateScheduledItem(ctx, repository.CreateScheduledItemInput{
		ID:                  uuid.New(),
		PetID:               p.PetID,
		SourceType:          sourceType,
		SourceID:            sourceID,
		Title:               title,
		Note:                note,
		StartsAt:            p.StartsAt.UTC(),
		PushEnabled:         pushEnabled,
		RemindOffsetMinutes: remindOffsetMinutes,
		RecurrenceRule:      recurrenceRule,
		RecurrenceInterval:  recurrenceInterval,
		RecurrenceUntil:     recurrenceUntil,
		CreatedBy:           p.UserID,
		UpdatedBy:           p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.regenerateScheduledItemOccurrences(ctx, item, time.Time{}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) UpdateScheduledItem(ctx context.Context, p UpdateScheduledItemParams) (*model.ScheduledItem, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ItemID == uuid.Nil || p.RowVersion <= 0 || p.StartsAt.IsZero() {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	current, err := s.repo.GetScheduledItem(ctx, p.PetID, p.ItemID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !isScheduledItemDirectlyWritable(current.SourceType) {
		return nil, ErrForbidden
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}
	note := trimOrNil(p.Note)
	recurrenceRule, recurrenceInterval, recurrenceUntil, err := validateScheduledRecurrence(p.RecurrenceRule, p.RecurrenceInterval, p.RecurrenceUntil)
	if err != nil {
		return nil, err
	}
	pushEnabled, remindOffsetMinutes, err := validateScheduledReminderSettings(p.PushEnabled, p.RemindOffsetMinutes)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.UpdateScheduledItem(ctx, repository.UpdateScheduledItemInput{
		ID:                  p.ItemID,
		PetID:               p.PetID,
		RowVersion:          p.RowVersion,
		Title:               title,
		Note:                note,
		StartsAt:            p.StartsAt.UTC(),
		PushEnabled:         pushEnabled,
		RemindOffsetMinutes: remindOffsetMinutes,
		RecurrenceRule:      recurrenceRule,
		RecurrenceInterval:  recurrenceInterval,
		RecurrenceUntil:     recurrenceUntil,
		UpdatedBy:           p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.regenerateScheduledItemOccurrences(ctx, item, time.Now().UTC()); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) DeleteScheduledItem(ctx context.Context, p DeleteScheduledItemParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ItemID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return err
	}
	current, err := s.repo.GetScheduledItem(ctx, p.PetID, p.ItemID)
	if err != nil {
		return mapRepoErr(err)
	}
	if !isScheduledItemDirectlyWritable(current.SourceType) {
		return ErrForbidden
	}
	return mapRepoErr(s.repo.DeleteScheduledItem(ctx, repository.DeleteScheduledItemInput{
		ID:         p.ItemID,
		PetID:      p.PetID,
		RowVersion: p.RowVersion,
		DeletedBy:  p.UserID,
	}))
}

func (s *Service) UpdateScheduledItemReminderSettings(ctx context.Context, p UpdateScheduledItemReminderSettingsParams) (*model.ScheduledItem, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ItemID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	if err := s.requireHealthWrite(ctx, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	_, err := s.repo.GetScheduledItem(ctx, p.PetID, p.ItemID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	remindOffsetMinutes, err := validateReminderOffset(p.RemindOffsetMinutes)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.UpdateScheduledItemReminderSettings(ctx, repository.UpdateScheduledItemReminderSettingsInput{
		ID:                  p.ItemID,
		PetID:               p.PetID,
		RowVersion:          p.RowVersion,
		PushEnabled:         p.PushEnabled,
		RemindOffsetMinutes: remindOffsetMinutes,
		UpdatedBy:           p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) GetScheduledItemOccurrence(ctx context.Context, userID, petID, occurrenceID uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error) {
	if userID == uuid.Nil || petID == uuid.Nil || occurrenceID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, petID, userID); err != nil {
		return nil, err
	}
	item, err := s.repo.GetScheduledItemOccurrence(ctx, petID, occurrenceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) ListScheduledItemOccurrences(ctx context.Context, p ListScheduledItemOccurrencesParams) (repository.ListScheduledItemOccurrencesOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListScheduledItemOccurrencesOutput{}, ErrInvalidInput
	}
	if _, err := s.requireHealthRead(ctx, p.PetID, p.UserID); err != nil {
		return repository.ListScheduledItemOccurrencesOutput{}, err
	}
	sourceType := normalizeEnumPtr(p.SourceType, []string{
		model.ScheduledItemSourceTypeManual,
		model.ScheduledItemSourceTypeLogType,
		model.ScheduledItemSourceTypePetEvent,
		model.ScheduledItemSourceTypeVetVisit,
		model.ScheduledItemSourceTypeVaccination,
		model.ScheduledItemSourceTypeProcedure,
	})
	if strings.TrimSpace(p.SourceType) != "" && sourceType == nil {
		return repository.ListScheduledItemOccurrencesOutput{}, ErrInvalidInput
	}
	out, err := s.repo.ListScheduledItemOccurrences(ctx, repository.ListScheduledItemOccurrencesInput{
		PetID:      p.PetID,
		Cursor:     p.Cursor,
		Limit:      p.Limit,
		DateFrom:   p.DateFrom,
		DateTo:     p.DateTo,
		SourceType: sourceType,
	})
	if err != nil {
		return repository.ListScheduledItemOccurrencesOutput{}, mapRepoErr(err)
	}
	return out, nil
}

func validateScheduledItemWritableSourceType(raw string) (string, error) {
	return validateEnum(strings.TrimSpace(strings.ToUpper(raw)), []string{
		model.ScheduledItemSourceTypeManual,
		model.ScheduledItemSourceTypeLogType,
		model.ScheduledItemSourceTypePetEvent,
	})
}

func validateScheduledItemSourceID(sourceType string, sourceID *uuid.UUID) (*uuid.UUID, error) {
	switch sourceType {
	case model.ScheduledItemSourceTypeLogType:
		if sourceID == nil || *sourceID == uuid.Nil {
			return nil, ErrInvalidInput
		}
		return sourceID, nil
	case model.ScheduledItemSourceTypeManual, model.ScheduledItemSourceTypePetEvent:
		if sourceID != nil && *sourceID == uuid.Nil {
			return nil, ErrInvalidInput
		}
		return sourceID, nil
	default:
		return nil, ErrInvalidInput
	}
}

func validateScheduledRecurrence(rule *string, interval *int, until *time.Time) (*string, *int, *time.Time, error) {
	if rule == nil {
		if interval != nil || until != nil {
			return nil, nil, nil, ErrInvalidInput
		}
		return nil, nil, nil, nil
	}
	normalizedRule := strings.TrimSpace(strings.ToUpper(*rule))
	if normalizedRule == "" {
		if interval != nil || until != nil {
			return nil, nil, nil, ErrInvalidInput
		}
		return nil, nil, nil, nil
	}
	validRule, err := validateEnum(normalizedRule, []string{
		model.RecurrenceRuleDaily,
		model.RecurrenceRuleWeekly,
		model.RecurrenceRuleMonthly,
		model.RecurrenceRuleYearly,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if interval == nil || *interval <= 0 {
		return nil, nil, nil, ErrInvalidInput
	}
	ruleCopy := validRule
	intervalCopy := *interval
	var untilCopy *time.Time
	if until != nil {
		if until.IsZero() {
			return nil, nil, nil, ErrInvalidInput
		}
		u := until.UTC()
		untilCopy = &u
	}
	return &ruleCopy, &intervalCopy, untilCopy, nil
}

func validateScheduledReminderSettings(pushEnabled *bool, remindOffsetMinutes *int) (bool, *int, error) {
	offset, err := validateReminderOffset(remindOffsetMinutes)
	if err != nil {
		return false, nil, err
	}
	enabled := true
	if pushEnabled != nil {
		enabled = *pushEnabled
	}
	if offset == nil {
		zero := 0
		offset = &zero
	}
	return enabled, offset, nil
}

func validateReminderOffset(remindOffsetMinutes *int) (*int, error) {
	if remindOffsetMinutes == nil {
		return nil, nil
	}
	if *remindOffsetMinutes < 0 {
		return nil, ErrInvalidInput
	}
	offset := *remindOffsetMinutes
	return &offset, nil
}

func isScheduledItemDirectlyWritable(sourceType string) bool {
	switch strings.ToUpper(strings.TrimSpace(sourceType)) {
	case model.ScheduledItemSourceTypeManual, model.ScheduledItemSourceTypeLogType, model.ScheduledItemSourceTypePetEvent:
		return true
	default:
		return false
	}
}

func (s *Service) regenerateScheduledItemOccurrences(ctx context.Context, item *model.ScheduledItem, from time.Time) error {
	if item == nil {
		return ErrInvalidInput
	}
	if !from.IsZero() {
		if err := s.repo.DeleteScheduledItemOccurrencesFrom(ctx, repository.DeleteScheduledItemOccurrencesFromInput{
			ScheduledItemID: item.ID,
			From:            from.UTC(),
		}); err != nil {
			return mapRepoErr(err)
		}
	}
	for _, scheduledFor := range buildScheduledOccurrences(item, from.UTC()) {
		if _, err := s.repo.CreateScheduledItemOccurrence(ctx, repository.CreateScheduledItemOccurrenceInput{
			ID:              uuid.New(),
			ScheduledItemID: item.ID,
			PetID:           item.PetID,
			ScheduledFor:    scheduledFor,
		}); err != nil && err != repository.ErrConflict {
			return mapRepoErr(err)
		}
	}
	return nil
}

func buildScheduledOccurrences(item *model.ScheduledItem, from time.Time) []time.Time {
	if item == nil || item.StartsAt.IsZero() {
		return []time.Time{}
	}
	if item.RecurrenceRule == nil {
		if from.IsZero() || !item.StartsAt.Before(from) {
			return []time.Time{item.StartsAt.UTC()}
		}
		return []time.Time{}
	}

	horizon := time.Now().UTC().AddDate(0, 0, scheduledOccurrencesHorizonDays)
	if item.RecurrenceUntil != nil && item.RecurrenceUntil.Before(horizon) {
		horizon = item.RecurrenceUntil.UTC()
	}
	next := item.StartsAt.UTC()
	if from.IsZero() {
		from = next
	}
	for next.Before(from) {
		var ok bool
		next, ok = nextScheduledTime(next, *item.RecurrenceRule, derefRecurrenceInterval(item.RecurrenceInterval))
		if !ok {
			return []time.Time{}
		}
	}
	out := make([]time.Time, 0)
	for !next.After(horizon) {
		out = append(out, next)
		var ok bool
		next, ok = nextScheduledTime(next, *item.RecurrenceRule, derefRecurrenceInterval(item.RecurrenceInterval))
		if !ok {
			break
		}
		if item.RecurrenceUntil != nil && next.After(item.RecurrenceUntil.UTC()) {
			break
		}
	}
	return out
}

func nextScheduledTime(current time.Time, rule string, interval int) (time.Time, bool) {
	switch rule {
	case model.RecurrenceRuleDaily:
		return current.AddDate(0, 0, interval), true
	case model.RecurrenceRuleWeekly:
		return current.AddDate(0, 0, 7*interval), true
	case model.RecurrenceRuleMonthly:
		return current.AddDate(0, interval, 0), true
	case model.RecurrenceRuleYearly:
		return current.AddDate(interval, 0, 0), true
	default:
		return time.Time{}, false
	}
}

func derefRecurrenceInterval(v *int) int {
	if v == nil || *v <= 0 {
		return 1
	}
	return *v
}

func (s *Service) syncSystemOneShotScheduledItem(ctx context.Context, petID uuid.UUID, sourceType string, sourceID uuid.UUID, title string, note *string, startsAt *time.Time, shouldExist bool, userID uuid.UUID) error {
	if !shouldExist || startsAt == nil || startsAt.IsZero() {
		return mapRepoErr(s.repo.DeleteHealthScheduledItem(ctx, repository.DeleteHealthScheduledItemInput{
			PetID:           petID,
			SourceType:      sourceType,
			SourceID:        sourceID,
			DeletedByUserID: userID,
		}))
	}
	item, err := s.repo.UpsertHealthScheduledItem(ctx, repository.UpsertHealthScheduledItemInput{
		PetID:               petID,
		SourceType:          sourceType,
		SourceID:            sourceID,
		Title:               strings.TrimSpace(title),
		Note:                trimOrNil(note),
		StartsAt:            startsAt.UTC(),
		PushEnabled:         false,
		RemindOffsetMinutes: nil,
		CreatedByUserID:     userID,
		UpdatedByUserID:     userID,
	})
	if err != nil {
		return mapRepoErr(err)
	}
	return s.regenerateScheduledItemOccurrences(ctx, item, time.Now().UTC())
}
