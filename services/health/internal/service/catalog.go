package service

import (
	"context"
	"health/internal/model"
	"health/internal/repository"
	"strings"

	"github.com/google/uuid"
)

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
	UnitCode  *string
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
	UnitCode   *string
	MinValue   *float64
	MaxValue   *float64
}

type DeleteMetricParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	MetricID   uuid.UUID
	RowVersion int
}

type GetBootstrapParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	IncludeCatalog bool
}

func (s *Service) GetLogsBootstrap(ctx context.Context, p GetBootstrapParams) (*model.LogComposerBootstrap, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	canWrite, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return nil, err
	}
	includeHealth, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionHealthRead)
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

	recent, err := s.repo.ListRecentLogTypes(ctx, p.PetID, includeHealth, 5)
	if err != nil {
		return nil, err
	}
	resp.RecentLogTypes = recent

	systemTypes, err := s.repo.ListLogTypes(ctx, repository.ListLogTypesInput{
		PetID:           p.PetID,
		Scope:           "SYSTEM",
		IncludeArchived: false,
	})
	if err != nil {
		return nil, err
	}
	customTypes, err := s.repo.ListLogTypes(ctx, repository.ListLogTypesInput{
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
		systemMetrics, err := s.repo.ListMetrics(ctx, repository.ListMetricsInput{
			PetID:           p.PetID,
			Scope:           "SYSTEM",
			IncludeArchived: false,
		})
		if err != nil {
			return nil, err
		}
		customMetrics, err := s.repo.ListMetrics(ctx, repository.ListMetricsInput{
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

func (s *Service) ListLogTypes(ctx context.Context, p ListLogTypesParams) ([]model.LogType, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return s.repo.ListLogTypes(ctx, repository.ListLogTypesInput{
		PetID:           p.PetID,
		Scope:           normalizeScope(p.Scope),
		Q:               strings.TrimSpace(p.Q),
		IncludeArchived: p.IncludeArchived,
		OnlyWithMetrics: p.OnlyWithMetrics,
	})
}

func (s *Service) CreateLogType(ctx context.Context, p CreateLogTypeParams) (*model.LogType, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
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
		metrics, err := s.repo.GetMetricsByIDs(ctx, p.PetID, metricIDs)
		if err != nil {
			return nil, err
		}
		if len(metrics) != len(metricIDs) {
			return nil, ErrInvalidInput
		}
	}

	item, err := s.repo.CreateLogType(ctx, repository.CreateLogTypeInput{
		ID:                 uuid.New(),
		PetID:              p.PetID,
		Name:               name,
		MetricRequirements: requirements,
		CreatedByUserID:    p.UserID,
		UpdatedByUserID:    p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) UpdateLogType(ctx context.Context, p UpdateLogTypeParams) (*model.LogType, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogTypeID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
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
		metrics, err := s.repo.GetMetricsByIDs(ctx, p.PetID, metricIDs)
		if err != nil {
			return nil, err
		}
		if len(metrics) != len(metricIDs) {
			return nil, ErrInvalidInput
		}
	}

	item, err := s.repo.UpdateLogType(ctx, repository.UpdateLogTypeInput{
		ID:                 p.LogTypeID,
		PetID:              p.PetID,
		RowVersion:         p.RowVersion,
		Name:               name,
		MetricRequirements: requirements,
		UpdatedByUserID:    p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) DeleteLogType(ctx context.Context, p DeleteLogTypeParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.LogTypeID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	err = s.repo.ArchiveLogType(ctx, repository.ArchiveLogTypeInput{
		ID:              p.LogTypeID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		DeletedByUserID: p.UserID,
	})
	if err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (s *Service) ListMetrics(ctx context.Context, p ListMetricsParams) ([]model.Metric, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return s.repo.ListMetrics(ctx, repository.ListMetricsInput{
		PetID:           p.PetID,
		Scope:           normalizeScope(p.Scope),
		Q:               strings.TrimSpace(p.Q),
		IncludeArchived: p.IncludeArchived,
		OnlyWithData:    p.OnlyWithData,
		OnlyWithUsage:   p.OnlyWithUsage,
	})
}

func (s *Service) CreateMetric(ctx context.Context, p CreateMetricParams) (*model.Metric, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
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
	unitCode := trimOrNil(p.UnitCode)
	minValue, maxValue, ok := normalizeRange(p.MinValue, p.MaxValue)
	if !ok {
		return nil, ErrInvalidInput
	}
	if inputKind == "SCALE" && (minValue == nil || maxValue == nil) {
		return nil, ErrInvalidInput
	}

	item, err := s.repo.CreateMetric(ctx, repository.CreateMetricInput{
		ID:              uuid.New(),
		PetID:           p.PetID,
		Name:            name,
		InputKind:       inputKind,
		UnitCode:        unitCode,
		MinValue:        minValue,
		MaxValue:        maxValue,
		CreatedByUserID: p.UserID,
		UpdatedByUserID: p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) UpdateMetric(ctx context.Context, p UpdateMetricParams) (*model.Metric, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.MetricID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
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
	unitCode := trimOrNil(p.UnitCode)
	minValue, maxValue, ok := normalizeRange(p.MinValue, p.MaxValue)
	if !ok {
		return nil, ErrInvalidInput
	}
	if inputKind == "SCALE" && (minValue == nil || maxValue == nil) {
		return nil, ErrInvalidInput
	}

	hasOutOfRange, err := s.repo.HasMetricValuesOutOfRange(ctx, p.PetID, p.MetricID, minValue, maxValue)
	if err != nil {
		return nil, err
	}
	if hasOutOfRange {
		return nil, ErrConflict
	}

	item, err := s.repo.UpdateMetric(ctx, repository.UpdateMetricInput{
		ID:              p.MetricID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		Name:            name,
		InputKind:       inputKind,
		UnitCode:        unitCode,
		MinValue:        minValue,
		MaxValue:        maxValue,
		UpdatedByUserID: p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return item, nil
}

func (s *Service) DeleteMetric(ctx context.Context, p DeleteMetricParams) error {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.MetricID == uuid.Nil || p.RowVersion <= 0 {
		return ErrInvalidInput
	}
	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	err = s.repo.ArchiveMetric(ctx, repository.ArchiveMetricInput{
		ID:              p.MetricID,
		PetID:           p.PetID,
		RowVersion:      p.RowVersion,
		DeletedByUserID: p.UserID,
	})
	if err != nil {
		return mapRepoErr(err)
	}
	return nil
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
	if v == "NUMERIC" || v == "SCALE" {
		return v
	}
	return ""
}

func normalizeRange(minValue, maxValue *float64) (*float64, *float64, bool) {
	if minValue == nil && maxValue == nil {
		return nil, nil, true
	}
	if minValue == nil || maxValue == nil {
		return nil, nil, false
	}
	if *minValue > *maxValue {
		return nil, nil, false
	}
	return minValue, maxValue, true
}

func normalizeRequirements(in []LogTypeRequirementInput) ([]repository.LogTypeMetricRequirementInput, []uuid.UUID, bool) {
	out := make([]repository.LogTypeMetricRequirementInput, 0, len(in))
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
		out = append(out, repository.LogTypeMetricRequirementInput{
			MetricID:   in[i].MetricID,
			IsRequired: in[i].IsRequired,
		})
		ids = append(ids, in[i].MetricID)
	}
	return out, ids, true
}
