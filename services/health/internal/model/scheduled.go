package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	ScheduledItemSourceTypeManual      = "MANUAL"
	ScheduledItemSourceTypeLogType     = "LOG_TYPE"
	ScheduledItemSourceTypePetEvent    = "PET_EVENT"
	ScheduledItemSourceTypeVetVisit    = "VET_VISIT"
	ScheduledItemSourceTypeVaccination = "VACCINATION"
	ScheduledItemSourceTypeProcedure   = "PROCEDURE"
)

const (
	ScheduledItemStatusActive    = "ACTIVE"
	ScheduledItemStatusDone      = "DONE"
	ScheduledItemStatusCancelled = "CANCELLED"
)

const (
	RecurrenceRuleDaily   = "DAILY"
	RecurrenceRuleWeekly  = "WEEKLY"
	RecurrenceRuleMonthly = "MONTHLY"
	RecurrenceRuleYearly  = "YEARLY"
)

type ScheduledItem struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	SourceType         string
	SourceID           *uuid.UUID
	Title              string
	Note               *string
	ScheduledFor       time.Time
	RecurrenceRule     *string
	RecurrenceInterval *int
	RecurrenceUntil    *time.Time
	Status             string
	RowVersion         int
	CreatedAt          time.Time
	CreatedByUserID    uuid.UUID
	UpdatedAt          time.Time
	UpdatedByUserID    uuid.UUID
	DeletedAt          *time.Time
	DeletedByUserID    *uuid.UUID
}

type ScheduledItemListItem struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	SourceType         string
	SourceID           *uuid.UUID
	Title              string
	NotePreview        *string
	ScheduledFor       time.Time
	RecurrenceRule     *string
	RecurrenceInterval *int
	RecurrenceUntil    *time.Time
	Status             string
	RowVersion         int
	CreatedAt          time.Time
	CreatedByUserID    uuid.UUID
	UpdatedAt          time.Time
	UpdatedByUserID    uuid.UUID
}

type ScheduledItemDispatch struct {
	ID              uuid.UUID
	ScheduledItemID uuid.UUID
	DispatchKey     string
	CreatedAt       time.Time
}

type CalendarPetInfo struct {
	PetID        uuid.UUID
	PetName      *string
	PetAvatarURL *string
}

type CalendarScheduledItem struct {
	ScheduledItem
	Pet CalendarPetInfo
}
