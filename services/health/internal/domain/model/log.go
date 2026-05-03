package model

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID                uuid.UUID
	PetID             uuid.UUID
	OccurredAt        time.Time
	LogTypeID         *uuid.UUID
	LogTypeName       *string
	LogTypeScope      *string
	Description       *string
	RelatedEntityType *string
	RelatedEntityID   *uuid.UUID
	RowVersion        int
	CreatedAt         time.Time
	CreatedByUserID   uuid.UUID
	UpdatedAt         time.Time
	UpdatedByUserID   uuid.UUID
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
	MetricValues      []LogMetricValue
	Attachments       []LogAttachment
}

type LogMetricValue struct {
	MetricID   uuid.UUID
	MetricName string
	InputKind  string
	Unit       *string
	ValueNum   float64
}

type LogAttachment struct {
	ID            uuid.UUID
	FileID        uuid.UUID
	FileName      *string
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
	RelatedEntityType  *string
	RelatedEntityID    *uuid.UUID
	RowVersion         int
	CreatedByUserID    uuid.UUID
	MetricValues       []LogMetricValue
	AttachmentsCount   int
	HasAttachments     bool
}

type Metric struct {
	ID              uuid.UUID
	Scope           string
	PetID           *uuid.UUID
	Code            *string
	Name            string
	InputKind       string
	Unit            *string
	MinValue        *float64
	MaxValue        *float64
	CreatedAt       time.Time
	CreatedByUserID *uuid.UUID
	UpdatedAt       time.Time
	UpdatedByUserID *uuid.UUID
	RowVersion      int
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
	Usage           MetricUsage
}

type LogType struct {
	ID                 uuid.UUID
	Scope              string
	PetID              *uuid.UUID
	Code               *string
	Name               string
	MetricRequirements []LogTypeMetricRequirement
	CreatedAt          time.Time
	CreatedByUserID    *uuid.UUID
	UpdatedAt          time.Time
	UpdatedByUserID    *uuid.UUID
	RowVersion         int
	DeletedAt          *time.Time
	DeletedByUserID    *uuid.UUID
}

type LogTypeMetricRequirement struct {
	MetricID   uuid.UUID
	IsRequired bool
}

type MetricUsage struct {
	LogTypesCount int
	LogsCount     int
}

type LogComposerBootstrap struct {
	Permissions    BootstrapPermissions
	RecentLogTypes []LogType
	SystemLogTypes []LogType
	CustomLogTypes []LogType
	SystemMetrics  []Metric
	CustomMetrics  []Metric
}

type BootstrapPermissions struct {
	LogRead  bool
	LogWrite bool
}

type HealthFile struct {
	ID       uuid.UUID
	MimeType string
	FileName *string
}

type AnalyticsMetricSummary struct {
	MetricID        uuid.UUID
	MetricName      string
	MetricScope     string
	InputKind       string
	Unit            *string
	PointsCount     int
	FirstOccurredAt time.Time
	LastOccurredAt  time.Time
	LastValueNum    float64
	UsedInLogTypes  []AnalyticsUsedLogType
}

type AnalyticsUsedLogType struct {
	LogTypeID   uuid.UUID
	LogTypeName string
}

type MetricSeriesPoint struct {
	OccurredAt  time.Time
	ValueNum    float64
	LogID       uuid.UUID
	LogTypeID   *uuid.UUID
	LogTypeName *string
}

type MetricSeriesSummary struct {
	PointsCount       int
	MinValueNum       float64
	MaxValueNum       float64
	LastValueNum      float64
	AvgValueNum       float64
	SumValueNum       float64
	DeltaFromFirstNum float64
}
