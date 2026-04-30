package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type TimeCursor struct {
	SortAt time.Time
	ID     uuid.UUID
}

type HealthAccessChecker interface {
	Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error)
}

type PetAccess struct {
	PetID       uuid.UUID
	PetRead     bool
	PetWrite    bool
	LogRead     bool
	LogWrite    bool
	HealthRead  bool
	HealthWrite bool
}

type HealthPetLister interface {
	ListPetAccessForUser(ctx context.Context, userID uuid.UUID) ([]PetAccess, error)
}

type ListScheduledItemsQuery struct {
	PetID       uuid.UUID
	Cursor      *TimeCursor
	Limit       int
	SourceType  *string
	SourceTypes []string
	DateFrom    *time.Time
	DateTo      *time.Time
	IncludePast bool
}

type ListScheduledItemsResult struct {
	Items      []model.ScheduledItemListItem
	NextCursor *TimeCursor
}

type ListRecurringScheduledItemsForHorizonParams struct {
	Now     time.Time
	Horizon time.Time
	Limit   int
}

type CreateScheduledItemInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	SourceType          string
	SourceID            *uuid.UUID
	Title               string
	Note                *string
	StartsAt            time.Time
	PushEnabled         bool
	RemindOffsetMinutes *int
	RecurrenceRule      *string
	RecurrenceInterval  *int
	RecurrenceUntil     *time.Time
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
}

type UpdateScheduledItemInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	RowVersion          int
	Title               string
	Note                *string
	StartsAt            time.Time
	PushEnabled         bool
	RemindOffsetMinutes *int
	RecurrenceRule      *string
	RecurrenceInterval  *int
	RecurrenceUntil     *time.Time
	UpdatedBy           uuid.UUID
}

type UpdateScheduledItemReminderSettingsInput struct {
	ID                  uuid.UUID
	PetID               uuid.UUID
	RowVersion          int
	PushEnabled         bool
	RemindOffsetMinutes *int
	UpdatedBy           uuid.UUID
}

type DeleteScheduledItemInput struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	DeletedBy  uuid.UUID
}

type ListScheduledItemOccurrencesQuery struct {
	PetID       uuid.UUID
	Cursor      *TimeCursor
	Limit       int
	DateFrom    *time.Time
	DateTo      *time.Time
	SourceType  *string
	SourceTypes []string
}

type ListScheduledItemOccurrencesResult struct {
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

type MarkScheduledItemOccurrencesGeneratedUntilInput struct {
	ScheduledItemID uuid.UUID
	GeneratedUntil  time.Time
}

type ScheduledRepository interface {
	GetScheduledItem(ctx context.Context, petID, itemID uuid.UUID) (*model.ScheduledItem, error)
	GetScheduledItemBySource(ctx context.Context, petID uuid.UUID, sourceType string, sourceID uuid.UUID) (*model.ScheduledItem, error)
	ListScheduledItems(ctx context.Context, query ListScheduledItemsQuery) (ListScheduledItemsResult, error)
	ListRecurringScheduledItemsForHorizon(ctx context.Context, params ListRecurringScheduledItemsForHorizonParams) ([]model.ScheduledItem, error)
	CreateScheduledItem(ctx context.Context, input CreateScheduledItemInput) (*model.ScheduledItem, error)
	UpdateScheduledItem(ctx context.Context, input UpdateScheduledItemInput) (*model.ScheduledItem, error)
	UpdateScheduledItemReminderSettings(ctx context.Context, input UpdateScheduledItemReminderSettingsInput) (*model.ScheduledItem, error)
	DeleteScheduledItem(ctx context.Context, input DeleteScheduledItemInput) error
	UpsertHealthScheduledItem(ctx context.Context, input UpsertHealthScheduledItemInput) (*model.ScheduledItem, error)
	DeleteHealthScheduledItem(ctx context.Context, input DeleteHealthScheduledItemInput) error
	GetScheduledItemOccurrence(ctx context.Context, petID, occurrenceID uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error)
	ListScheduledItemOccurrences(ctx context.Context, query ListScheduledItemOccurrencesQuery) (ListScheduledItemOccurrencesResult, error)
	ListCalendarDayScheduledOccurrences(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error)
	ListCalendarDayScheduledOccurrencesForPets(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error)
	CreateScheduledItemOccurrence(ctx context.Context, input CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error)
	DeleteScheduledItemOccurrencesFrom(ctx context.Context, input DeleteScheduledItemOccurrencesFromInput) error
	MarkScheduledItemOccurrencesGeneratedUntil(ctx context.Context, input MarkScheduledItemOccurrencesGeneratedUntilInput) error
}

type UpsertHealthScheduledItemInput struct {
	PetID               uuid.UUID
	SourceType          string
	SourceID            uuid.UUID
	Title               string
	Note                *string
	StartsAt            time.Time
	PushEnabled         bool
	RemindOffsetMinutes *int
	CreatedByUserID     uuid.UUID
	UpdatedByUserID     uuid.UUID
}

type DeleteHealthScheduledItemInput struct {
	PetID           uuid.UUID
	SourceType      string
	SourceID        uuid.UUID
	DeletedByUserID uuid.UUID
}
