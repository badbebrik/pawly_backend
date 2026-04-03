package repository

import (
	"context"
	"health/internal/model"
	"time"

	"github.com/google/uuid"
)

type ListCursor struct {
	OccurredAt time.Time
	ID         uuid.UUID
}

type ListLogsInput struct {
	PetID           uuid.UUID
	Cursor          *ListCursor
	Limit           int
	Sort            string
	Q               string
	DateFrom        *time.Time
	DateTo          *time.Time
	TypeIDs         []uuid.UUID
	Source          *string
	HasAttachments  *bool
	HasMetricValues *bool
}

type ListLogsOutput struct {
	Items      []model.LogListItem
	NextCursor *ListCursor
}

type LogMetricValueInput struct {
	MetricID uuid.UUID
	ValueNum float64
}

type CreateLogInput struct {
	ID                uuid.UUID
	PetID             uuid.UUID
	OccurredAt        time.Time
	LogTypeID         *uuid.UUID
	Description       *string
	Source            string
	CreatedByUserID   uuid.UUID
	UpdatedByUserID   uuid.UUID
	MetricValues      []LogMetricValueInput
	AttachmentFileIDs []uuid.UUID
}

type UpdateLogInput struct {
	ID                uuid.UUID
	PetID             uuid.UUID
	RowVersion        int
	OccurredAt        time.Time
	LogTypeID         *uuid.UUID
	Description       *string
	UpdatedByUserID   uuid.UUID
	MetricValues      []LogMetricValueInput
	AttachmentFileIDs []uuid.UUID
}

type DeleteLogInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RowVersion      int
	DeletedByUserID uuid.UUID
}

type UpsertHealthEntityLogInput struct {
	PetID           uuid.UUID
	EntityType      string
	EntityID        uuid.UUID
	OccurredAt      time.Time
	Description     *string
	CreatedByUserID uuid.UUID
	UpdatedByUserID uuid.UUID
}

type DeleteHealthEntityLogInput struct {
	PetID           uuid.UUID
	EntityType      string
	EntityID        uuid.UUID
	DeletedByUserID uuid.UUID
}

type ListLogTypesInput struct {
	PetID           uuid.UUID
	Scope           string
	Q               string
	IncludeArchived bool
	OnlyWithMetrics bool
}

type LogTypeMetricRequirementInput struct {
	MetricID   uuid.UUID
	IsRequired bool
}

type CreateLogTypeInput struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	Name               string
	MetricRequirements []LogTypeMetricRequirementInput
	CreatedByUserID    uuid.UUID
	UpdatedByUserID    uuid.UUID
}

type UpdateLogTypeInput struct {
	ID                 uuid.UUID
	PetID              uuid.UUID
	RowVersion         int
	Name               string
	MetricRequirements []LogTypeMetricRequirementInput
	UpdatedByUserID    uuid.UUID
}

type ArchiveLogTypeInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RowVersion      int
	DeletedByUserID uuid.UUID
}

type ListMetricsInput struct {
	PetID           uuid.UUID
	Scope           string
	Q               string
	IncludeArchived bool
	OnlyWithData    bool
	OnlyWithUsage   bool
}

type CreateMetricInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	Name            string
	InputKind       string
	Unit            *string
	MinValue        *float64
	MaxValue        *float64
	CreatedByUserID uuid.UUID
	UpdatedByUserID uuid.UUID
}

type UpdateMetricInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RowVersion      int
	Name            string
	InputKind       string
	Unit            *string
	MinValue        *float64
	MaxValue        *float64
	UpdatedByUserID uuid.UUID
}

type ArchiveMetricInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RowVersion      int
	DeletedByUserID uuid.UUID
}

type ListAnalyticsMetricsInput struct {
	PetID    uuid.UUID
	Q        string
	DateFrom *time.Time
	DateTo   *time.Time
	Source   *string
	TypeIDs  []uuid.UUID
	Limit    int
}

type ListMetricSeriesInput struct {
	PetID          uuid.UUID
	MetricID       uuid.UUID
	DateFrom       *time.Time
	DateTo         *time.Time
	Source         *string
	TypeIDs        []uuid.UUID
	Sort           string
	IncludeSummary bool
}

type LogRepository interface {
	GetLog(ctx context.Context, petID, logID uuid.UUID) (*model.Log, error)
	ListLogs(ctx context.Context, in ListLogsInput) (ListLogsOutput, error)
	CreateLog(ctx context.Context, in CreateLogInput) (*model.Log, error)
	UpdateLog(ctx context.Context, in UpdateLogInput) (*model.Log, error)
	SoftDeleteLog(ctx context.Context, in DeleteLogInput) error
	UpsertHealthEntityLog(ctx context.Context, in UpsertHealthEntityLogInput) error
	DeleteHealthEntityLog(ctx context.Context, in DeleteHealthEntityLogInput) error
	GetLogTypeByID(ctx context.Context, petID uuid.UUID, logTypeID uuid.UUID) (*model.LogType, error)
	GetMetricsByIDs(ctx context.Context, petID uuid.UUID, metricIDs []uuid.UUID) (map[uuid.UUID]model.Metric, error)
	ListLogTypes(ctx context.Context, in ListLogTypesInput) ([]model.LogType, error)
	CreateLogType(ctx context.Context, in CreateLogTypeInput) (*model.LogType, error)
	UpdateLogType(ctx context.Context, in UpdateLogTypeInput) (*model.LogType, error)
	ArchiveLogType(ctx context.Context, in ArchiveLogTypeInput) error
	ListRecentLogTypes(ctx context.Context, petID uuid.UUID, includeHealth bool, limit int) ([]model.LogType, error)
	ListMetrics(ctx context.Context, in ListMetricsInput) ([]model.Metric, error)
	GetMetricByID(ctx context.Context, petID, metricID uuid.UUID) (*model.Metric, error)
	CreateMetric(ctx context.Context, in CreateMetricInput) (*model.Metric, error)
	UpdateMetric(ctx context.Context, in UpdateMetricInput) (*model.Metric, error)
	ArchiveMetric(ctx context.Context, in ArchiveMetricInput) error
	HasMetricValuesOutOfRange(ctx context.Context, petID, metricID uuid.UUID, minValue, maxValue *float64) (bool, error)
	ListAnalyticsMetrics(ctx context.Context, in ListAnalyticsMetricsInput) ([]model.AnalyticsMetricSummary, error)
	ListMetricSeries(ctx context.Context, in ListMetricSeriesInput) ([]model.MetricSeriesPoint, *model.MetricSeriesSummary, error)
}
