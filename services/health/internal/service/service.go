package service

import (
	"context"
	"health/internal/model"
	"health/internal/repository"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ActionLogRead     = "log_read"
	ActionLogWrite    = "log_write"
	ActionHealthRead  = "health_read"
	ActionHealthWrite = "health_write"
)

type ACLClient interface {
	Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error)
}

type FileClient interface {
	EnsureFilesExist(ctx context.Context, fileIDs []uuid.UUID) error
	BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetFiles(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]model.HealthFile, error)
	LinkAttachments(ctx context.Context, petID uuid.UUID, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error
	UnlinkAttachments(ctx context.Context, entityType string, entityID uuid.UUID, fileIDs []uuid.UUID) error
}

type Service struct {
	repo repository.Repository
	acl  ACLClient
	file FileClient
}

func New(repo repository.Repository, acl ACLClient, file FileClient) *Service {
	return &Service{repo: repo, acl: acl, file: file}
}

type ListLogsParams struct {
	UserID          uuid.UUID
	PetID           uuid.UUID
	Cursor          *repository.ListCursor
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

type CreateOrUpdateMetricValue struct {
	MetricID uuid.UUID
	ValueNum float64
}

type CreateLogParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	OccurredAt        time.Time
	LogTypeID         *uuid.UUID
	Description       *string
	MetricValues      []CreateOrUpdateMetricValue
	AttachmentFileIDs []uuid.UUID
}

type UpdateLogParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	LogID             uuid.UUID
	RowVersion        int
	OccurredAt        time.Time
	LogTypeID         *uuid.UUID
	Description       *string
	MetricValues      []CreateOrUpdateMetricValue
	AttachmentFileIDs []uuid.UUID
}

type DeleteLogParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	LogID      uuid.UUID
	RowVersion int
}

func (s *Service) ListLogs(ctx context.Context, p ListLogsParams) (repository.ListLogsOutput, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return repository.ListLogsOutput{}, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return repository.ListLogsOutput{}, err
	}
	if !allowed {
		return repository.ListLogsOutput{}, ErrForbidden
	}

	healthAllowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionHealthRead)
	if err != nil {
		return repository.ListLogsOutput{}, err
	}

	source := normalizeSource(p.Source)
	if !healthAllowed {
		if source != nil && *source == "HEALTH" {
			return repository.ListLogsOutput{Items: []model.LogListItem{}}, nil
		}
		userSource := "USER"
		source = &userSource
	}

	return s.repo.ListLogs(ctx, repository.ListLogsInput{
		PetID:           p.PetID,
		Cursor:          p.Cursor,
		Limit:           p.Limit,
		Sort:            p.Sort,
		Q:               strings.TrimSpace(p.Q),
		DateFrom:        p.DateFrom,
		DateTo:          p.DateTo,
		TypeIDs:         uniqueUUIDs(p.TypeIDs),
		Source:          source,
		HasAttachments:  p.HasAttachments,
		HasMetricValues: p.HasMetricValues,
	})
}

func (s *Service) GetLog(ctx context.Context, userID, petID, logID uuid.UUID) (*model.Log, error) {
	if userID == uuid.Nil || petID == uuid.Nil || logID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, petID, userID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	item, err := s.repo.GetLog(ctx, petID, logID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if item.Source == "HEALTH" {
		healthAllowed, err := s.acl.Check(ctx, petID, userID, ActionHealthRead)
		if err != nil {
			return nil, err
		}
		if !healthAllowed {
			return nil, ErrNotFound
		}
	}

	s.enrichAttachmentURLs(ctx, item)
	return item, nil
}

func (s *Service) CreateLog(ctx context.Context, p CreateLogParams) (*model.Log, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.OccurredAt.IsZero() {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	cleanMetricValues, cleanAttachments, logTypeID, description, err := s.validateLogPayload(ctx, p.PetID, p.LogTypeID, p.Description, p.MetricValues, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	item, err := s.repo.CreateLog(ctx, repository.CreateLogInput{
		ID:                id,
		PetID:             p.PetID,
		OccurredAt:        p.OccurredAt,
		LogTypeID:         logTypeID,
		Description:       description,
		Source:            "USER",
		CreatedByUserID:   p.UserID,
		UpdatedByUserID:   p.UserID,
		MetricValues:      cleanMetricValues,
		AttachmentFileIDs: cleanAttachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}

	if len(cleanAttachments) > 0 {
		if err := s.file.LinkAttachments(ctx, p.PetID, "LOG", id, cleanAttachments); err != nil {
			return nil, err
		}
	}

	s.enrichAttachmentURLs(ctx, item)
	return item, nil
}

func (s *Service) UpdateLog(ctx context.Context, p UpdateLogParams) (*model.Log, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogID == uuid.Nil || p.RowVersion <= 0 || p.OccurredAt.IsZero() {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	cleanMetricValues, cleanAttachments, logTypeID, description, err := s.validateLogPayload(ctx, p.PetID, p.LogTypeID, p.Description, p.MetricValues, p.AttachmentFileIDs)
	if err != nil {
		return nil, err
	}
	current, err := s.repo.GetLog(ctx, p.PetID, p.LogID)
	if err != nil {
		return nil, mapRepoErr(err)
	}

	item, err := s.repo.UpdateLog(ctx, repository.UpdateLogInput{
		ID:                p.LogID,
		PetID:             p.PetID,
		RowVersion:        p.RowVersion,
		OccurredAt:        p.OccurredAt,
		LogTypeID:         logTypeID,
		Description:       description,
		UpdatedByUserID:   p.UserID,
		MetricValues:      cleanMetricValues,
		AttachmentFileIDs: cleanAttachments,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}

	addIDs, removeIDs := diffFileIDs(logAttachmentFileIDs(current.Attachments), cleanAttachments)
	if len(addIDs) > 0 {
		if err := s.file.LinkAttachments(ctx, p.PetID, "LOG", p.LogID, addIDs); err != nil {
			return nil, err
		}
	}
	if len(removeIDs) > 0 {
		if err := s.file.UnlinkAttachments(ctx, "LOG", p.LogID, removeIDs); err != nil {
			return nil, err
		}
	}

	s.enrichAttachmentURLs(ctx, item)
	return item, nil
}

func (s *Service) DeleteLog(ctx context.Context, p DeleteLogParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	current, err := s.repo.GetLog(ctx, p.PetID, p.LogID)
	if err != nil {
		return mapRepoErr(err)
	}

	err = s.repo.SoftDeleteLog(ctx, repository.DeleteLogInput{
		ID:              p.LogID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		DeletedByUserID: p.UserID,
	})
	if err != nil {
		return mapRepoErr(err)
	}
	fileIDs := logAttachmentFileIDs(current.Attachments)
	if len(fileIDs) > 0 {
		if err := s.file.UnlinkAttachments(ctx, "LOG", p.LogID, fileIDs); err != nil {
			return err
		}
	}
	return nil
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

func healthAttachmentFileIDs(items []model.HealthAttachment) []uuid.UUID {
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

func (s *Service) validateLogPayload(
	ctx context.Context,
	petID uuid.UUID,
	logTypeID *uuid.UUID,
	description *string,
	metricValues []CreateOrUpdateMetricValue,
	attachmentFileIDs []uuid.UUID,
) ([]repository.LogMetricValueInput, []uuid.UUID, *uuid.UUID, *string, error) {
	cleanDescription := trimOrNil(description)
	cleanAttachments := uniqueUUIDs(attachmentFileIDs)
	cleanMetricValues := make([]repository.LogMetricValueInput, 0, len(metricValues))
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
		cleanMetricValues = append(cleanMetricValues, repository.LogMetricValueInput{
			MetricID: metricValues[i].MetricID,
			ValueNum: metricValues[i].ValueNum,
		})
		metricIDs = append(metricIDs, metricValues[i].MetricID)
	}

	if len(cleanAttachments) > 0 {
		if err := s.file.EnsureFilesExist(ctx, cleanAttachments); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	metricsByID, err := s.repo.GetMetricsByIDs(ctx, petID, metricIDs)
	if err != nil {
		return nil, nil, nil, nil, mapRepoErr(err)
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
		logType, err := s.repo.GetLogTypeByID(ctx, petID, *logTypeID)
		if err != nil {
			if err == repository.ErrNotFound {
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
		return cleanMetricValues, cleanAttachments, logTypeID, cleanDescription, nil
	}

	return cleanMetricValues, cleanAttachments, nil, cleanDescription, nil
}

func (s *Service) enrichAttachmentURLs(ctx context.Context, item *model.Log) {
	if item == nil || len(item.Attachments) == 0 {
		return
	}
	fileIDs := make([]uuid.UUID, 0, len(item.Attachments))
	for i := range item.Attachments {
		fileIDs = append(fileIDs, item.Attachments[i].FileID)
	}
	urls, err := s.file.BatchGetDownloadURLs(ctx, fileIDs)
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

func mapRepoErr(err error) error {
	switch err {
	case repository.ErrNotFound:
		return ErrNotFound
	case repository.ErrConflict:
		return ErrConflict
	default:
		return err
	}
}

func normalizeSource(raw *string) *string {
	if raw == nil {
		return nil
	}
	v := strings.ToUpper(strings.TrimSpace(*raw))
	if v != "USER" && v != "HEALTH" {
		return nil
	}
	return &v
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

func trimOrNil(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
