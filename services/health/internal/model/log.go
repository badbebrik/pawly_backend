package model

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID               uuid.UUID
	PetID            uuid.UUID
	OccurredAt       time.Time
	LogTypeID        *uuid.UUID
	LogTypeName      *string
	LogTypeScope     *string
	Description      *string
	Source           string
	SourceEntityType *string
	SourceEntityID   *uuid.UUID
	RowVersion       int
	CreatedAt        time.Time
	CreatedByUserID  uuid.UUID
	UpdatedAt        time.Time
	UpdatedByUserID  uuid.UUID
	DeletedAt        *time.Time
	DeletedByUserID  *uuid.UUID
	MetricValues     []LogMetricValue
	Attachments      []LogAttachment
}

type LogMetricValue struct {
	MetricID   uuid.UUID
	MetricName string
	InputKind  string
	UnitCode   *string
	ValueNum   float64
}

type LogAttachment struct {
	ID            uuid.UUID
	FileID        uuid.UUID
	FileType      string
	AddedByUserID uuid.UUID
	AddedAt       time.Time
	DownloadURL   *string
	PreviewURL    *string
}

type LogListItem struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	OccurredAt         time.Time
	LogTypeID          *uuid.UUID
	LogTypeName        *string
	LogTypeScope       *string
	DescriptionPreview *string
	Source             string
	SourceEntityType   *string
	SourceEntityID     *uuid.UUID
	RowVersion         int
	CreatedByUserID    uuid.UUID
	AttachmentsCount   int
	HasAttachments     bool
}

type Metric struct {
	ID        uuid.UUID
	Scope     string
	PetID     *uuid.UUID
	Code      *string
	Name      string
	InputKind string
	UnitCode  *string
	MinValue  *float64
	MaxValue  *float64
}

type LogType struct {
	ID                 uuid.UUID
	Scope              string
	PetID              *uuid.UUID
	Code               *string
	Name               string
	MetricRequirements []LogTypeMetricRequirement
}

type LogTypeMetricRequirement struct {
	MetricID   uuid.UUID `json:"metric_id"`
	IsRequired bool      `json:"is_required"`
}
