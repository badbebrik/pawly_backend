package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ActionLogRead  = "log_read"
	ActionLogWrite = "log_write"
)

type Logs struct {
	repo ports.LogsRepository
	acl  ports.HealthAccessChecker
	file ports.HealthFileClient
}

func NewLogs(repo ports.LogsRepository, acl ports.HealthAccessChecker, file ports.HealthFileClient) *Logs {
	return &Logs{repo: repo, acl: acl, file: file}
}

type ListLogsParams struct {
	UserID          uuid.UUID
	PetID           uuid.UUID
	Cursor          *ports.LogCursor
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
	NextCursor *ports.LogCursor
}

type CreateOrUpdateMetricValue struct {
	MetricID uuid.UUID
	ValueNum float64
}

type CreateLogParams struct {
	UserID       uuid.UUID
	PetID        uuid.UUID
	OccurredAt   time.Time
	LogTypeID    *uuid.UUID
	Description  *string
	MetricValues []CreateOrUpdateMetricValue
	Attachments  []AttachmentParam
}

type UpdateLogParams struct {
	UserID       uuid.UUID
	PetID        uuid.UUID
	LogID        uuid.UUID
	RowVersion   int
	OccurredAt   time.Time
	LogTypeID    *uuid.UUID
	Description  *string
	MetricValues []CreateOrUpdateMetricValue
	Attachments  []AttachmentParam
}

type DeleteLogParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	LogID      uuid.UUID
	RowVersion int
}

type ListLogTypesParams struct {
	UserID          uuid.UUID
	PetID           uuid.UUID
	Scope           string
	Q               string
	IncludeArchived bool
	OnlyWithMetrics bool
}

type LogTypeRequirementInput struct {
	MetricID   uuid.UUID
	IsRequired bool
}

type CreateLogTypeParams struct {
	UserID             uuid.UUID
	PetID              uuid.UUID
	Name               string
	MetricRequirements []LogTypeRequirementInput
}

type UpdateLogTypeParams struct {
	UserID             uuid.UUID
	PetID              uuid.UUID
	LogTypeID          uuid.UUID
	RowVersion         int
	Name               string
	MetricRequirements []LogTypeRequirementInput
}

type DeleteLogTypeParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	LogTypeID  uuid.UUID
	RowVersion int
}

type ListMetricsParams struct {
	UserID          uuid.UUID
	PetID           uuid.UUID
	Scope           string
	Q               string
	IncludeArchived bool
	OnlyWithData    bool
	OnlyWithUsage   bool
}

type CreateMetricParams struct {
	UserID    uuid.UUID
	PetID     uuid.UUID
	Name      string
	InputKind string
	Unit      *string
	MinValue  *float64
	MaxValue  *float64
}

type UpdateMetricParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	MetricID   uuid.UUID
	RowVersion int
	Name       string
	InputKind  string
	Unit       *string
	MinValue   *float64
	MaxValue   *float64
}

type DeleteMetricParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	MetricID   uuid.UUID
	RowVersion int
}

type GetLogsBootstrapParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	IncludeCatalog bool
}

type InitAttachmentUploadParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	EntityType        *string
	MimeType          string
	OriginalFilename  string
	ExpectedSizeBytes int64
}

type ConfirmAttachmentUploadParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	EntityType *string
	FileID     uuid.UUID
	SizeBytes  int64
}

func (u *Logs) InitAttachmentUpload(ctx context.Context, p InitAttachmentUploadParams) (uuid.UUID, ports.UploadInfo, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || strings.TrimSpace(p.MimeType) == "" || p.ExpectedSizeBytes <= 0 {
		return uuid.Nil, ports.UploadInfo{}, ErrInvalidInput
	}
	if err := u.requireAttachmentUploadWrite(ctx, p.PetID, p.UserID, p.EntityType); err != nil {
		return uuid.Nil, ports.UploadInfo{}, err
	}
	return u.file.InitUpload(ctx, strings.TrimSpace(strings.ToLower(p.MimeType)), p.ExpectedSizeBytes, strings.TrimSpace(p.OriginalFilename))
}

func (u *Logs) ConfirmAttachmentUpload(ctx context.Context, p ConfirmAttachmentUploadParams) (*ports.UploadedFile, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.FileID == uuid.Nil || p.SizeBytes <= 0 {
		return nil, ErrInvalidInput
	}
	if err := u.requireAttachmentUploadWrite(ctx, p.PetID, p.UserID, p.EntityType); err != nil {
		return nil, err
	}
	return u.file.ConfirmUpload(ctx, p.FileID, p.SizeBytes)
}

func (u *Logs) requireAttachmentUploadWrite(ctx context.Context, petID, userID uuid.UUID, entityTypeInput *string) error {
	if entityTypeInput != nil && strings.TrimSpace(*entityTypeInput) == "" {
		entityTypeInput = nil
	}
	entityType := normalizeHealthDocumentEntityType(entityTypeInput)
	if entityTypeInput != nil && entityType == nil {
		return ErrInvalidInput
	}
	action := ActionHealthWrite
	if entityType != nil && *entityType == "LOG" {
		action = ActionLogWrite
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

func (u *Logs) GetLogsBootstrap(ctx context.Context, p GetLogsBootstrapParams) (*model.LogComposerBootstrap, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	canWrite, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	resp := &model.LogComposerBootstrap{
		Permissions: model.BootstrapPermissions{
			LogRead:  true,
			LogWrite: canWrite,
		},
		RecentLogTypes: []model.LogType{},
		SystemLogTypes: []model.LogType{},
		CustomLogTypes: []model.LogType{},
		SystemMetrics:  []model.Metric{},
		CustomMetrics:  []model.Metric{},
	}

	recent, err := u.repo.ListRecentLogTypes(ctx, p.PetID, 5)
	if err != nil {
		return nil, err
	}
	resp.RecentLogTypes = recent

	systemTypes, err := u.repo.ListLogTypes(ctx, ports.ListLogTypesInput{
		PetID:           p.PetID,
		Scope:           "SYSTEM",
		IncludeArchived: false,
	})
	if err != nil {
		return nil, err
	}
	customTypes, err := u.repo.ListLogTypes(ctx, ports.ListLogTypesInput{
		PetID:           p.PetID,
		Scope:           "CUSTOM",
		IncludeArchived: false,
	})
	if err != nil {
		return nil, err
	}
	resp.SystemLogTypes = systemTypes
	resp.CustomLogTypes = customTypes

	if p.IncludeCatalog {
		systemMetrics, err := u.repo.ListMetrics(ctx, ports.ListMetricsInput{
			PetID:           p.PetID,
			Scope:           "SYSTEM",
			IncludeArchived: false,
		})
		if err != nil {
			return nil, err
		}
		customMetrics, err := u.repo.ListMetrics(ctx, ports.ListMetricsInput{
			PetID:           p.PetID,
			Scope:           "CUSTOM",
			IncludeArchived: false,
		})
		if err != nil {
			return nil, err
		}
		resp.SystemMetrics = systemMetrics
		resp.CustomMetrics = customMetrics
	}

	return resp, nil
}

func (u *Logs) ListLogs(ctx context.Context, p ListLogsParams) (ListLogsResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListLogsResult{}, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return ListLogsResult{}, err
	}
	if !allowed {
		return ListLogsResult{}, ErrForbidden
	}
	out, err := u.repo.ListLogs(ctx, ports.ListLogsQuery{
		PetID:           p.PetID,
		Cursor:          p.Cursor,
		Limit:           p.Limit,
		Sort:            p.Sort,
		Q:               strings.TrimSpace(p.Q),
		DateFrom:        p.DateFrom,
		DateTo:          p.DateTo,
		TypeIDs:         uniqueUUIDs(p.TypeIDs),
		HasAttachments:  p.HasAttachments,
		HasMetricValues: p.HasMetricValues,
	})
	if err != nil {
		return ListLogsResult{}, mapLogRepoErr(err)
	}
	return ListLogsResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *Logs) ListLogTypes(ctx context.Context, p ListLogTypesParams) ([]model.LogType, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return u.repo.ListLogTypes(ctx, ports.ListLogTypesInput{
		PetID:           p.PetID,
		Scope:           normalizeScope(p.Scope),
		Q:               strings.TrimSpace(p.Q),
		IncludeArchived: p.IncludeArchived,
		OnlyWithMetrics: p.OnlyWithMetrics,
	})
}

func (u *Logs) CreateLogType(ctx context.Context, p CreateLogTypeParams) (*model.LogType, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	requirements, metricIDs, ok := normalizeRequirements(p.MetricRequirements)
	if !ok {
		return nil, ErrInvalidInput
	}
	if len(metricIDs) > 0 {
		metrics, err := u.repo.GetMetricsByIDs(ctx, p.PetID, metricIDs)
		if err != nil {
			return nil, err
		}
		if len(metrics) != len(metricIDs) {
			return nil, ErrInvalidInput
		}
	}
	item, err := u.repo.CreateLogType(ctx, ports.CreateLogTypeInput{
		ID:                 uuid.New(),
		PetID:              p.PetID,
		Name:               name,
		MetricRequirements: requirements,
		CreatedByUserID:    p.UserID,
		UpdatedByUserID:    p.UserID,
	})
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	return item, nil
}

func (u *Logs) UpdateLogType(ctx context.Context, p UpdateLogTypeParams) (*model.LogType, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogTypeID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	requirements, metricIDs, ok := normalizeRequirements(p.MetricRequirements)
	if !ok {
		return nil, ErrInvalidInput
	}
	if len(metricIDs) > 0 {
		metrics, err := u.repo.GetMetricsByIDs(ctx, p.PetID, metricIDs)
		if err != nil {
			return nil, err
		}
		if len(metrics) != len(metricIDs) {
			return nil, ErrInvalidInput
		}
	}
	item, err := u.repo.UpdateLogType(ctx, ports.UpdateLogTypeInput{
		ID:                 p.LogTypeID,
		PetID:              p.PetID,
		RowVersion:         p.RowVersion,
		Name:               name,
		MetricRequirements: requirements,
		UpdatedByUserID:    p.UserID,
	})
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	return item, nil
}

func (u *Logs) DeleteLogType(ctx context.Context, p DeleteLogTypeParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogTypeID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return mapLogRepoErr(u.repo.ArchiveLogType(ctx, ports.ArchiveLogTypeInput{
		ID:              p.LogTypeID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		DeletedByUserID: p.UserID,
	}))
}

func (u *Logs) ListMetrics(ctx context.Context, p ListMetricsParams) ([]model.Metric, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return u.repo.ListMetrics(ctx, ports.ListMetricsInput{
		PetID:           p.PetID,
		Scope:           normalizeScope(p.Scope),
		Q:               strings.TrimSpace(p.Q),
		IncludeArchived: p.IncludeArchived,
		OnlyWithData:    p.OnlyWithData,
		OnlyWithUsage:   p.OnlyWithUsage,
	})
}

func (u *Logs) CreateMetric(ctx context.Context, p CreateMetricParams) (*model.Metric, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	inputKind := normalizeInputKind(p.InputKind)
	if inputKind == "" {
		return nil, ErrInvalidInput
	}
	unit := trimStringOrNil(p.Unit)
	minValue, maxValue, ok := normalizeRange(p.MinValue, p.MaxValue)
	if !ok {
		return nil, ErrInvalidInput
	}
	if inputKind == "SCALE" && (minValue == nil || maxValue == nil) {
		return nil, ErrInvalidInput
	}
	if inputKind == "BOOLEAN" && (unit != nil || minValue != nil || maxValue != nil) {
		return nil, ErrInvalidInput
	}
	item, err := u.repo.CreateMetric(ctx, ports.CreateMetricInput{
		ID:              uuid.New(),
		PetID:           p.PetID,
		Name:            name,
		InputKind:       inputKind,
		Unit:            unit,
		MinValue:        minValue,
		MaxValue:        maxValue,
		CreatedByUserID: p.UserID,
		UpdatedByUserID: p.UserID,
	})
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	return item, nil
}

func (u *Logs) UpdateMetric(ctx context.Context, p UpdateMetricParams) (*model.Metric, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.MetricID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	inputKind := normalizeInputKind(p.InputKind)
	if inputKind == "" {
		return nil, ErrInvalidInput
	}
	unit := trimStringOrNil(p.Unit)
	minValue, maxValue, ok := normalizeRange(p.MinValue, p.MaxValue)
	if !ok {
		return nil, ErrInvalidInput
	}
	if inputKind == "SCALE" && (minValue == nil || maxValue == nil) {
		return nil, ErrInvalidInput
	}
	if inputKind == "BOOLEAN" && (unit != nil || minValue != nil || maxValue != nil) {
		return nil, ErrInvalidInput
	}
	hasOutOfRange, err := u.repo.HasMetricValuesOutOfRange(ctx, p.PetID, p.MetricID, minValue, maxValue)
	if err != nil {
		return nil, err
	}
	if hasOutOfRange {
		return nil, ErrConflict
	}
	item, err := u.repo.UpdateMetric(ctx, ports.UpdateMetricInput{
		ID:              p.MetricID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		Name:            name,
		InputKind:       inputKind,
		Unit:            unit,
		MinValue:        minValue,
		MaxValue:        maxValue,
		UpdatedByUserID: p.UserID,
	})
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	return item, nil
}

func (u *Logs) DeleteMetric(ctx context.Context, p DeleteMetricParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.MetricID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return mapLogRepoErr(u.repo.ArchiveMetric(ctx, ports.ArchiveMetricInput{
		ID:              p.MetricID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		DeletedByUserID: p.UserID,
	}))
}

func (u *Logs) GetLog(ctx context.Context, userID, petID, logID uuid.UUID) (*model.Log, error) {
	if userID == uuid.Nil || petID == uuid.Nil || logID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	item, err := u.repo.GetLog(ctx, petID, logID)
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	u.enrichAttachmentURLs(ctx, item)
	return item, nil
}

func (u *Logs) CreateLog(ctx context.Context, p CreateLogParams) (*model.Log, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.OccurredAt.IsZero() {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	cleanMetricValues, cleanAttachments, logTypeID, description, err := u.validateLogPayload(ctx, p.PetID, p.LogTypeID, p.Description, p.MetricValues, p.Attachments)
	if err != nil {
		return nil, err
	}
	id := uuid.New()
	item, err := u.repo.CreateLog(ctx, ports.CreateLogInput{
		ID:              id,
		PetID:           p.PetID,
		OccurredAt:      p.OccurredAt,
		LogTypeID:       logTypeID,
		Description:     description,
		CreatedByUserID: p.UserID,
		UpdatedByUserID: p.UserID,
		MetricValues:    cleanMetricValues,
		Attachments:     cleanAttachments,
	})
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	if len(cleanAttachments) > 0 {
		if err := u.file.LinkAttachments(ctx, p.PetID, "LOG", id, attachmentInputsFileIDs(cleanAttachments)); err != nil {
			return nil, err
		}
	}
	u.enrichAttachmentURLs(ctx, item)
	return item, nil
}

func (u *Logs) UpdateLog(ctx context.Context, p UpdateLogParams) (*model.Log, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogID == uuid.Nil || p.RowVersion <= 0 || p.OccurredAt.IsZero() {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	cleanMetricValues, cleanAttachments, logTypeID, description, err := u.validateLogPayload(ctx, p.PetID, p.LogTypeID, p.Description, p.MetricValues, p.Attachments)
	if err != nil {
		return nil, err
	}
	current, err := u.repo.GetLog(ctx, p.PetID, p.LogID)
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	item, err := u.repo.UpdateLog(ctx, ports.UpdateLogInput{
		ID:              p.LogID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		OccurredAt:      p.OccurredAt,
		LogTypeID:       logTypeID,
		Description:     description,
		UpdatedByUserID: p.UserID,
		MetricValues:    cleanMetricValues,
		Attachments:     cleanAttachments,
	})
	if err != nil {
		return nil, mapLogRepoErr(err)
	}
	addIDs, removeIDs := diffFileIDs(logAttachmentFileIDs(current.Attachments), attachmentInputsFileIDs(cleanAttachments))
	if len(addIDs) > 0 {
		if err := u.file.LinkAttachments(ctx, p.PetID, "LOG", p.LogID, addIDs); err != nil {
			return nil, err
		}
	}
	if len(removeIDs) > 0 {
		if err := u.file.UnlinkAttachments(ctx, "LOG", p.LogID, removeIDs); err != nil {
			return nil, err
		}
		if err := u.file.DeleteFilesIfUnlinked(ctx, removeIDs); err != nil {
			return nil, err
		}
	}
	u.enrichAttachmentURLs(ctx, item)
	return item, nil
}

func (u *Logs) DeleteLog(ctx context.Context, p DeleteLogParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	current, err := u.repo.GetLog(ctx, p.PetID, p.LogID)
	if err != nil {
		return mapLogRepoErr(err)
	}
	if err := u.repo.SoftDeleteLog(ctx, ports.DeleteLogInput{
		ID:              p.LogID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		DeletedByUserID: p.UserID,
	}); err != nil {
		return mapLogRepoErr(err)
	}
	fileIDs := logAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := u.file.UnlinkAttachments(ctx, "LOG", p.LogID, fileIDs); err != nil {
			return err
		}
		if err := u.file.DeleteFilesIfUnlinked(ctx, fileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (u *Logs) validateLogPayload(ctx context.Context, petID uuid.UUID, logTypeID *uuid.UUID, description *string, metricValues []CreateOrUpdateMetricValue, attachments []AttachmentParam) ([]ports.LogMetricValueInput, []ports.AttachmentInput, *uuid.UUID, *string, error) {
	cleanDescription := trimStringOrNil(description)
	cleanMetricValues := make([]ports.LogMetricValueInput, 0, len(metricValues))
	seenMetric := make(map[uuid.UUID]struct{}, len(metricValues))
	metricIDs := make([]uuid.UUID, 0, len(metricValues))
	for i := range metricValues {
		if metricValues[i].MetricID == uuid.Nil {
			return nil, nil, nil, nil, ErrInvalidInput
		}
		if _, exists := seenMetric[metricValues[i].MetricID]; exists {
			return nil, nil, nil, nil, ErrInvalidInput
		}
		seenMetric[metricValues[i].MetricID] = struct{}{}
		cleanMetricValues = append(cleanMetricValues, ports.LogMetricValueInput{
			MetricID: metricValues[i].MetricID,
			ValueNum: metricValues[i].ValueNum,
		})
		metricIDs = append(metricIDs, metricValues[i].MetricID)
	}

	var preparedAttachments []ports.AttachmentInput
	if len(attachments) > 0 {
		attachments, err := u.prepareHealthAttachments(ctx, attachments)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		preparedAttachments = attachments
	} else {
		preparedAttachments = []ports.AttachmentInput{}
	}

	metricsByID, err := u.repo.GetMetricsByIDs(ctx, petID, metricIDs)
	if err != nil {
		return nil, nil, nil, nil, mapLogRepoErr(err)
	}
	if len(metricsByID) != len(metricIDs) {
		return nil, nil, nil, nil, ErrInvalidInput
	}
	for i := range cleanMetricValues {
		metric := metricsByID[cleanMetricValues[i].MetricID]
		if metric.InputKind == "BOOLEAN" && cleanMetricValues[i].ValueNum != 0 && cleanMetricValues[i].ValueNum != 1 {
			return nil, nil, nil, nil, ErrInvalidInput
		}
		if metric.InputKind == "SCALE" && math.Trunc(cleanMetricValues[i].ValueNum) != cleanMetricValues[i].ValueNum {
			return nil, nil, nil, nil, ErrInvalidInput
		}
		if metric.MinValue != nil && cleanMetricValues[i].ValueNum < *metric.MinValue {
			return nil, nil, nil, nil, ErrInvalidInput
		}
		if metric.MaxValue != nil && cleanMetricValues[i].ValueNum > *metric.MaxValue {
			return nil, nil, nil, nil, ErrInvalidInput
		}
	}

	if logTypeID != nil {
		if *logTypeID == uuid.Nil {
			return nil, nil, nil, nil, ErrInvalidInput
		}
		logType, err := u.repo.GetLogTypeByID(ctx, petID, *logTypeID)
		if err != nil {
			if err == ports.ErrNotFound {
				return nil, nil, nil, nil, ErrInvalidInput
			}
			return nil, nil, nil, nil, err
		}
		required := make(map[uuid.UUID]struct{})
		for i := range logType.MetricRequirements {
			if logType.MetricRequirements[i].IsRequired {
				required[logType.MetricRequirements[i].MetricID] = struct{}{}
			}
		}
		for metricID := range required {
			if _, ok := seenMetric[metricID]; !ok {
				return nil, nil, nil, nil, ErrInvalidInput
			}
		}
		return cleanMetricValues, preparedAttachments, logTypeID, cleanDescription, nil
	}

	return cleanMetricValues, preparedAttachments, nil, cleanDescription, nil
}

func (u *Logs) prepareHealthAttachments(ctx context.Context, attachments []AttachmentParam) ([]ports.AttachmentInput, error) {
	return prepareHealthAttachments(ctx, u.file, attachments)
}

func (u *Logs) enrichAttachmentURLs(ctx context.Context, item *model.Log) {
	if item == nil || len(item.Attachments) == 0 {
		return
	}
	fileIDs := make([]uuid.UUID, 0, len(item.Attachments))
	for i := range item.Attachments {
		fileIDs = append(fileIDs, item.Attachments[i].FileID)
	}
	urls, err := u.file.BatchGetDownloadURLs(ctx, fileIDs)
	if err != nil {
		return
	}
	for i := range item.Attachments {
		if u, ok := urls[item.Attachments[i].FileID]; ok && strings.TrimSpace(u) != "" {
			urlCopy := u
			item.Attachments[i].DownloadURL = &urlCopy
			if item.Attachments[i].FileType == "image" {
				item.Attachments[i].PreviewURL = &urlCopy
			}
		}
	}
}

func mapLogRepoErr(err error) error {
	switch err {
	case ports.ErrNotFound:
		return ErrNotFound
	case ports.ErrConflict:
		return ErrConflict
	default:
		return err
	}
}

func uniqueUUIDs(items []uuid.UUID) []uuid.UUID {
	if len(items) == 0 {
		return []uuid.UUID{}
	}
	seen := make(map[uuid.UUID]struct{}, len(items))
	out := make([]uuid.UUID, 0, len(items))
	for i := range items {
		if items[i] == uuid.Nil {
			continue
		}
		if _, ok := seen[items[i]]; ok {
			continue
		}
		seen[items[i]] = struct{}{}
		out = append(out, items[i])
	}
	return out
}

func trimStringOrNil(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func attachmentInputsFileIDs(items []ports.AttachmentInput) []uuid.UUID {
	if len(items) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(items))
	for i := range items {
		out = append(out, items[i].FileID)
	}
	return out
}

func logAttachmentFileIDs(items []model.LogAttachment) []uuid.UUID {
	if len(items) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(items))
	for i := range items {
		out = append(out, items[i].FileID)
	}
	return out
}

func diffFileIDs(existing []uuid.UUID, desired []uuid.UUID) ([]uuid.UUID, []uuid.UUID) {
	existingSet := make(map[uuid.UUID]struct{}, len(existing))
	for i := range existing {
		existingSet[existing[i]] = struct{}{}
	}
	desiredSet := make(map[uuid.UUID]struct{}, len(desired))
	for i := range desired {
		desiredSet[desired[i]] = struct{}{}
	}
	add := make([]uuid.UUID, 0)
	for i := range desired {
		if _, ok := existingSet[desired[i]]; ok {
			continue
		}
		add = append(add, desired[i])
	}
	remove := make([]uuid.UUID, 0)
	for i := range existing {
		if _, ok := desiredSet[existing[i]]; ok {
			continue
		}
		remove = append(remove, existing[i])
	}
	return add, remove
}

func detectAttachmentFileType(mimeType string) string {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	default:
		return "file"
	}
}

func normalizeScope(raw string) string {
	scope := strings.ToUpper(strings.TrimSpace(raw))
	if scope == "SYSTEM" || scope == "CUSTOM" {
		return scope
	}
	return "ALL"
}

func normalizeInputKind(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	if v == "NUMERIC" || v == "SCALE" || v == "BOOLEAN" {
		return v
	}
	return ""
}

func normalizeRange(minValue, maxValue *float64) (*float64, *float64, bool) {
	if minValue == nil && maxValue == nil {
		return nil, nil, true
	}
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return nil, nil, false
	}
	return minValue, maxValue, true
}

func normalizeRequirements(in []LogTypeRequirementInput) ([]ports.LogTypeMetricRequirementInput, []uuid.UUID, bool) {
	out := make([]ports.LogTypeMetricRequirementInput, 0, len(in))
	ids := make([]uuid.UUID, 0, len(in))
	seen := make(map[uuid.UUID]struct{}, len(in))
	for i := range in {
		if in[i].MetricID == uuid.Nil {
			return nil, nil, false
		}
		if _, ok := seen[in[i].MetricID]; ok {
			return nil, nil, false
		}
		seen[in[i].MetricID] = struct{}{}
		out = append(out, ports.LogTypeMetricRequirementInput{
			MetricID:   in[i].MetricID,
			IsRequired: in[i].IsRequired,
		})
		ids = append(ids, in[i].MetricID)
	}
	return out, ids, true
}
