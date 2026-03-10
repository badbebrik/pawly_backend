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
	SourceEntityType  *string
	SourceEntityID    *uuid.UUID
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

type LogRepository interface {
	GetLog(ctx context.Context, petID, logID uuid.UUID) (*model.Log, error)
	ListLogs(ctx context.Context, in ListLogsInput) (ListLogsOutput, error)
	CreateLog(ctx context.Context, in CreateLogInput) (*model.Log, error)
	UpdateLog(ctx context.Context, in UpdateLogInput) (*model.Log, error)
	SoftDeleteLog(ctx context.Context, in DeleteLogInput) error
	GetLogTypeByID(ctx context.Context, petID uuid.UUID, logTypeID uuid.UUID) (*model.LogType, error)
	GetMetricsByIDs(ctx context.Context, petID uuid.UUID, metricIDs []uuid.UUID) (map[uuid.UUID]model.Metric, error)
}
