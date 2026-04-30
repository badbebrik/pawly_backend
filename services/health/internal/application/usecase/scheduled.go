package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ActionPetRead     = "pet_read"
	ActionPetWrite    = "pet_write"
	ActionHealthRead  = "health_read"
	ActionHealthWrite = "health_write"
)

const scheduledOccurrencesHorizonDays = 90

type Scheduled struct {
	repo ports.ScheduledRepository
	acl  ports.HealthAccessChecker
}

func NewScheduled(repo ports.ScheduledRepository, acl ports.HealthAccessChecker) *Scheduled {
	return &Scheduled{repo: repo, acl: acl}
}

type GenerateScheduledOccurrencesHorizonResult struct {
	Scanned int
	Failed  int
}

type ListScheduledItemsParams struct {
	UserID      uuid.UUID
	PetID       uuid.UUID
	Cursor      *ports.TimeCursor
	Limit       int
	SourceType  string
	DateFrom    *time.Time
	DateTo      *time.Time
	IncludePast bool
}

type ListScheduledItemsResult struct {
	Items      []model.ScheduledItemListItem
	NextCursor *ports.TimeCursor
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
	Cursor     *ports.TimeCursor
	Limit      int
	SourceType string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type ListScheduledItemOccurrencesResult struct {
	Items      []model.ScheduledItemOccurrenceListItem
	NextCursor *ports.TimeCursor
}

func (u *Scheduled) GetScheduledItem(ctx context.Context, userID, petID, itemID uuid.UUID) (*model.ScheduledItem, error) {
	if userID == uuid.Nil || petID == uuid.Nil || itemID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	item, err := u.repo.GetScheduledItem(ctx, petID, itemID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := u.requireScheduledSourceRead(ctx, petID, userID, item.SourceType); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *Scheduled) ListScheduledItems(ctx context.Context, p ListScheduledItemsParams) (ListScheduledItemsResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListScheduledItemsResult{}, ErrInvalidInput
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
		return ListScheduledItemsResult{}, ErrInvalidInput
	}
	sourceTypes, err := u.allowedScheduledReadSourceTypes(ctx, p.PetID, p.UserID, sourceType)
	if err != nil {
		return ListScheduledItemsResult{}, err
	}
	out, err := u.repo.ListScheduledItems(ctx, ports.ListScheduledItemsQuery{
		PetID:       p.PetID,
		Cursor:      p.Cursor,
		Limit:       p.Limit,
		SourceType:  sourceType,
		SourceTypes: sourceTypes,
		DateFrom:    p.DateFrom,
		DateTo:      p.DateTo,
		IncludePast: p.IncludePast,
	})
	if err != nil {
		return ListScheduledItemsResult{}, mapRepoErr(err)
	}
	return ListScheduledItemsResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *Scheduled) CreateScheduledItem(ctx context.Context, p CreateScheduledItemParams) (*model.ScheduledItem, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.StartsAt.IsZero() {
		return nil, ErrInvalidInput
	}
	sourceType, err := validateScheduledItemWritableSourceType(p.SourceType)
	if err != nil {
		return nil, err
	}
	if err := u.requireScheduledSourceWrite(ctx, p.PetID, p.UserID, sourceType); err != nil {
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
	if recurrenceUntil != nil && recurrenceUntil.Before(p.StartsAt.UTC()) {
		return nil, ErrInvalidInput
	}
	pushEnabled, remindOffsetMinutes, err := validateScheduledReminderSettings(p.PushEnabled, p.RemindOffsetMinutes)
	if err != nil {
		return nil, err
	}
	item, err := u.repo.CreateScheduledItem(ctx, ports.CreateScheduledItemInput{
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
	if err := u.regenerateScheduledItemOccurrences(ctx, item, time.Now().UTC()); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *Scheduled) UpdateScheduledItem(ctx context.Context, p UpdateScheduledItemParams) (*model.ScheduledItem, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ItemID == uuid.Nil || p.RowVersion <= 0 || p.StartsAt.IsZero() {
		return nil, ErrInvalidInput
	}
	current, err := u.repo.GetScheduledItem(ctx, p.PetID, p.ItemID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !isScheduledItemDirectlyWritable(current.SourceType) {
		return nil, ErrForbidden
	}
	if err := u.requireScheduledSourceWrite(ctx, p.PetID, p.UserID, current.SourceType); err != nil {
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
	if recurrenceUntil != nil && recurrenceUntil.Before(p.StartsAt.UTC()) {
		return nil, ErrInvalidInput
	}
	pushEnabled, remindOffsetMinutes, err := validateScheduledReminderSettings(p.PushEnabled, p.RemindOffsetMinutes)
	if err != nil {
		return nil, err
	}
	item, err := u.repo.UpdateScheduledItem(ctx, ports.UpdateScheduledItemInput{
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
	if err := u.regenerateScheduledItemOccurrences(ctx, item, time.Now().UTC()); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *Scheduled) DeleteScheduledItem(ctx context.Context, p DeleteScheduledItemParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ItemID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	current, err := u.repo.GetScheduledItem(ctx, p.PetID, p.ItemID)
	if err != nil {
		return mapRepoErr(err)
	}
	if !isScheduledItemDirectlyWritable(current.SourceType) {
		return ErrForbidden
	}
	if err := u.requireScheduledSourceWrite(ctx, p.PetID, p.UserID, current.SourceType); err != nil {
		return err
	}
	return mapRepoErr(u.repo.DeleteScheduledItem(ctx, ports.DeleteScheduledItemInput{
		ID:         p.ItemID,
		PetID:      p.PetID,
		RowVersion: p.RowVersion,
		DeletedBy:  p.UserID,
	}))
}

func (u *Scheduled) UpdateScheduledItemReminderSettings(ctx context.Context, p UpdateScheduledItemReminderSettingsParams) (*model.ScheduledItem, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.ItemID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	current, err := u.repo.GetScheduledItem(ctx, p.PetID, p.ItemID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := u.requireScheduledSourceWrite(ctx, p.PetID, p.UserID, current.SourceType); err != nil {
		return nil, err
	}
	remindOffsetMinutes, err := validateReminderOffset(p.RemindOffsetMinutes)
	if err != nil {
		return nil, err
	}
	item, err := u.repo.UpdateScheduledItemReminderSettings(ctx, ports.UpdateScheduledItemReminderSettingsInput{
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

func (u *Scheduled) GetScheduledItemOccurrence(ctx context.Context, userID, petID, occurrenceID uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error) {
	if userID == uuid.Nil || petID == uuid.Nil || occurrenceID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	item, err := u.repo.GetScheduledItemOccurrence(ctx, petID, occurrenceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if err := u.requireScheduledSourceRead(ctx, petID, userID, item.Rule.SourceType); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *Scheduled) ListScheduledItemOccurrences(ctx context.Context, p ListScheduledItemOccurrencesParams) (ListScheduledItemOccurrencesResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListScheduledItemOccurrencesResult{}, ErrInvalidInput
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
		return ListScheduledItemOccurrencesResult{}, ErrInvalidInput
	}
	sourceTypes, err := u.allowedScheduledReadSourceTypes(ctx, p.PetID, p.UserID, sourceType)
	if err != nil {
		return ListScheduledItemOccurrencesResult{}, err
	}
	out, err := u.repo.ListScheduledItemOccurrences(ctx, ports.ListScheduledItemOccurrencesQuery{
		PetID:       p.PetID,
		Cursor:      p.Cursor,
		Limit:       p.Limit,
		DateFrom:    p.DateFrom,
		DateTo:      p.DateTo,
		SourceType:  sourceType,
		SourceTypes: sourceTypes,
	})
	if err != nil {
		return ListScheduledItemOccurrencesResult{}, mapRepoErr(err)
	}
	return ListScheduledItemOccurrencesResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *Scheduled) GenerateOccurrencesHorizon(ctx context.Context, limit int) (GenerateScheduledOccurrencesHorizonResult, error) {
	if limit <= 0 {
		limit = 100
	}
	now := time.Now().UTC()
	items, err := u.repo.ListRecurringScheduledItemsForHorizon(ctx, ports.ListRecurringScheduledItemsForHorizonParams{
		Now:     now,
		Horizon: scheduledOccurrencesHorizon(now),
		Limit:   limit,
	})
	if err != nil {
		return GenerateScheduledOccurrencesHorizonResult{}, mapRepoErr(err)
	}
	result := GenerateScheduledOccurrencesHorizonResult{Scanned: len(items)}
	for i := range items {
		if err := ensureScheduledItemOccurrences(ctx, u.repo, &items[i], now); err != nil {
			result.Failed++
		}
	}
	return result, nil
}

func (u *Scheduled) allowedScheduledReadSourceTypes(ctx context.Context, petID, userID uuid.UUID, requestedSourceType *string) ([]string, error) {
	if requestedSourceType != nil {
		if err := u.requireScheduledSourceRead(ctx, petID, userID, *requestedSourceType); err != nil {
			return nil, err
		}
		return nil, nil
	}
	access, err := u.getScheduledReadAccess(ctx, petID, userID)
	if err != nil {
		return nil, err
	}
	sourceTypes := scheduledReadSourceTypes(access)
	if len(sourceTypes) == 0 {
		return nil, ErrForbidden
	}
	return sourceTypes, nil
}

func (u *Scheduled) requireScheduledSourceRead(ctx context.Context, petID, userID uuid.UUID, sourceType string) error {
	action, ok := scheduledSourceReadAction(sourceType)
	if !ok {
		return ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, petID, userID, action)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

func (u *Scheduled) requireScheduledSourceWrite(ctx context.Context, petID, userID uuid.UUID, sourceType string) error {
	action, ok := scheduledSourceWriteAction(sourceType)
	if !ok {
		return ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, petID, userID, action)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

func (u *Scheduled) getScheduledReadAccess(ctx context.Context, petID, userID uuid.UUID) (ports.PetAccess, error) {
	var access ports.PetAccess
	access.PetID = petID
	checks := []struct {
		action string
		set    func(bool)
	}{
		{ActionPetRead, func(v bool) { access.PetRead = v }},
		{ActionLogRead, func(v bool) { access.LogRead = v }},
		{ActionHealthRead, func(v bool) { access.HealthRead = v }},
	}
	for i := range checks {
		allowed, err := u.acl.Check(ctx, petID, userID, checks[i].action)
		if err != nil {
			return ports.PetAccess{}, err
		}
		checks[i].set(allowed)
	}
	return access, nil
}

func scheduledReadSourceTypes(access ports.PetAccess) []string {
	sourceTypes := make([]string, 0, 6)
	if access.PetRead {
		sourceTypes = append(sourceTypes, model.ScheduledItemSourceTypeManual, model.ScheduledItemSourceTypePetEvent)
	}
	if access.LogRead {
		sourceTypes = append(sourceTypes, model.ScheduledItemSourceTypeLogType)
	}
	if access.HealthRead {
		sourceTypes = append(sourceTypes,
			model.ScheduledItemSourceTypeVetVisit,
			model.ScheduledItemSourceTypeVaccination,
			model.ScheduledItemSourceTypeProcedure,
		)
	}
	return sourceTypes
}

func scheduledSourceReadAction(sourceType string) (string, bool) {
	switch sourceType {
	case model.ScheduledItemSourceTypeManual, model.ScheduledItemSourceTypePetEvent:
		return ActionPetRead, true
	case model.ScheduledItemSourceTypeLogType:
		return ActionLogRead, true
	case model.ScheduledItemSourceTypeVetVisit, model.ScheduledItemSourceTypeVaccination, model.ScheduledItemSourceTypeProcedure:
		return ActionHealthRead, true
	default:
		return "", false
	}
}

func scheduledSourceWriteAction(sourceType string) (string, bool) {
	switch sourceType {
	case model.ScheduledItemSourceTypeManual, model.ScheduledItemSourceTypePetEvent:
		return ActionPetWrite, true
	case model.ScheduledItemSourceTypeLogType:
		return ActionLogWrite, true
	case model.ScheduledItemSourceTypeVetVisit, model.ScheduledItemSourceTypeVaccination, model.ScheduledItemSourceTypeProcedure:
		return ActionHealthWrite, true
	default:
		return "", false
	}
}

func (u *Scheduled) requireHealthRead(ctx context.Context, petID, userID uuid.UUID) (bool, error) {
	allowed, err := u.acl.Check(ctx, petID, userID, ActionHealthRead)
	if err != nil {
		return false, err
	}
	if !allowed {
		return false, ErrForbidden
	}
	return true, nil
}

func (u *Scheduled) requireHealthWrite(ctx context.Context, petID, userID uuid.UUID) error {
	allowed, err := u.acl.Check(ctx, petID, userID, ActionHealthWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
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

func (u *Scheduled) regenerateScheduledItemOccurrences(ctx context.Context, item *model.ScheduledItem, from time.Time) error {
	if item == nil {
		return ErrInvalidInput
	}
	if !from.IsZero() {
		if err := u.repo.DeleteScheduledItemOccurrencesFrom(ctx, ports.DeleteScheduledItemOccurrencesFromInput{
			ScheduledItemID: item.ID,
			From:            from.UTC(),
		}); err != nil {
			return mapRepoErr(err)
		}
	}
	return ensureScheduledItemOccurrences(ctx, u.repo, item, from.UTC())
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

	horizon := scheduledOccurrencesHorizon(time.Now().UTC())
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

func ensureScheduledItemOccurrences(ctx context.Context, repo ports.ScheduledRepository, item *model.ScheduledItem, from time.Time) error {
	if item == nil {
		return ErrInvalidInput
	}
	horizon := scheduledOccurrencesHorizon(time.Now().UTC())
	for _, scheduledFor := range buildScheduledOccurrences(item, from.UTC()) {
		if _, err := repo.CreateScheduledItemOccurrence(ctx, ports.CreateScheduledItemOccurrenceInput{
			ID:              uuid.New(),
			ScheduledItemID: item.ID,
			PetID:           item.PetID,
			ScheduledFor:    scheduledFor,
		}); err != nil && err != ports.ErrConflict {
			return mapRepoErr(err)
		}
	}
	if item.RecurrenceRule != nil {
		return mapRepoErr(repo.MarkScheduledItemOccurrencesGeneratedUntil(ctx, ports.MarkScheduledItemOccurrencesGeneratedUntilInput{
			ScheduledItemID: item.ID,
			GeneratedUntil:  horizon,
		}))
	}
	return nil
}

func scheduledOccurrencesHorizon(now time.Time) time.Time {
	return now.UTC().AddDate(0, 0, scheduledOccurrencesHorizonDays)
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

func mapRepoErr(err error) error {
	switch err {
	case nil:
		return nil
	case ports.ErrNotFound:
		return ErrNotFound
	case ports.ErrConflict:
		return ErrConflict
	default:
		return err
	}
}

func validateEnum(value string, allowed []string) (string, error) {
	for _, candidate := range allowed {
		if value == candidate {
			return value, nil
		}
	}
	return "", ErrInvalidInput
}

func normalizeEnumPtr(raw string, allowed []string) *string {
	trimmed := strings.TrimSpace(strings.ToUpper(raw))
	if trimmed == "" {
		return nil
	}
	for _, candidate := range allowed {
		if trimmed == candidate {
			value := candidate
			return &value
		}
	}
	return nil
}

func trimOrNil(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
