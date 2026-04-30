package ports

import (
	"context"
	"health/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type LogCursor struct {
	OccurredAt time.Time
	ID         uuid.UUID
}

type ListLogsQuery struct {
	PetID           uuid.UUID
	Cursor          *LogCursor
	Limit           int
	Sort            string
	Q               string
	DateFrom        *time.Time
	DateTo          *time.Time
	TypeIDs         []uuid.UUID
	HasAttachments  *bool
	HasMetricValues *bool
}

type ListLogsResult struct {
	Items      []model.LogListItem
	NextCursor *LogCursor
}

type LogMetricValueInput struct {
	MetricID uuid.UUID
	ValueNum float64
}

type AttachmentInput struct {
	FileID   uuid.UUID
	FileName *string
	FileType string
}

type CreateLogInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	OccurredAt      time.Time
	LogTypeID       *uuid.UUID
	Description     *string
	CreatedByUserID uuid.UUID
	UpdatedByUserID uuid.UUID
	MetricValues    []LogMetricValueInput
	Attachments     []AttachmentInput
}

type UpdateLogInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RowVersion      int
	OccurredAt      time.Time
	LogTypeID       *uuid.UUID
	Description     *string
	UpdatedByUserID uuid.UUID
	MetricValues    []LogMetricValueInput
	Attachments     []AttachmentInput
}

type DeleteLogInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	RowVersion      int
	DeletedByUserID uuid.UUID
}

type UploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

type UploadedFile struct {
	ID               uuid.UUID
	MimeType         string
	SizeBytes        int64
	OriginalFilename *string
}

type LogsRepository interface {
	GetLog(ctx context.Context, petID, logID uuid.UUID) (*model.Log, error)
	ListLogs(ctx context.Context, query ListLogsQuery) (ListLogsResult, error)
	CreateLog(ctx context.Context, input CreateLogInput) (*model.Log, error)
	UpdateLog(ctx context.Context, input UpdateLogInput) (*model.Log, error)
	SoftDeleteLog(ctx context.Context, input DeleteLogInput) error
	GetLogTypeByID(ctx context.Context, petID, logTypeID uuid.UUID) (*model.LogType, error)
	GetMetricsByIDs(ctx context.Context, petID uuid.UUID, metricIDs []uuid.UUID) (map[uuid.UUID]model.Metric, error)
	ListRecentLogTypes(ctx context.Context, petID uuid.UUID, limit int) ([]model.LogType, error)
	ListLogTypes(ctx context.Context, input ListLogTypesInput) ([]model.LogType, error)
	CreateLogType(ctx context.Context, input CreateLogTypeInput) (*model.LogType, error)
	UpdateLogType(ctx context.Context, input UpdateLogTypeInput) (*model.LogType, error)
	ArchiveLogType(ctx context.Context, input ArchiveLogTypeInput) error
	ListMetrics(ctx context.Context, input ListMetricsInput) ([]model.Metric, error)
	CreateMetric(ctx context.Context, input CreateMetricInput) (*model.Metric, error)
	UpdateMetric(ctx context.Context, input UpdateMetricInput) (*model.Metric, error)
	ArchiveMetric(ctx context.Context, input ArchiveMetricInput) error
	HasMetricValuesOutOfRange(ctx context.Context, petID, metricID uuid.UUID, minValue, maxValue *float64) (bool, error)
	GetMetricByID(ctx context.Context, petID, metricID uuid.UUID) (*model.Metric, error)
	ListAnalyticsMetrics(ctx context.Context, input ListAnalyticsMetricsInput) ([]model.AnalyticsMetricSummary, error)
	ListMetricSeries(ctx context.Context, input ListMetricSeriesInput) ([]model.MetricSeriesPoint, *model.MetricSeriesSummary, error)
}

type HealthFileClient interface {
	InitUpload(ctx context.Context, mimeType string, expectedSize int64, originalFilename string) (uuid.UUID, UploadInfo, error)
	ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) (*UploadedFile, error)
	BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetFiles(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]model.HealthFile, error)
	LinkAttachments(ctx context.Context, petID uuid.UUID, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error
	UnlinkAttachments(ctx context.Context, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error
	DeleteFilesIfUnlinked(ctx context.Context, fileIDs []uuid.UUID) error
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
	TypeIDs  []uuid.UUID
	Limit    int
}

type ListMetricSeriesInput struct {
	PetID          uuid.UUID
	MetricID       uuid.UUID
	DateFrom       *time.Time
	DateTo         *time.Time
	TypeIDs        []uuid.UUID
	Sort           string
	IncludeSummary bool
}
