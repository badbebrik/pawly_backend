package service

import (
	"context"
	"health/internal/model"
	"health/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ListAnalyticsMetricsParams struct {
	UserID   uuid.UUID
	PetID    uuid.UUID
	Q        string
	DateFrom *time.Time
	DateTo   *time.Time
	Source   string
	TypeIDs  []uuid.UUID
	Limit    int
}

type GetMetricSeriesParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	MetricID       uuid.UUID
	DateFrom       *time.Time
	DateTo         *time.Time
	Source         string
	TypeIDs        []uuid.UUID
	Sort           string
	IncludeSummary bool
}

type MetricSeriesResult struct {
	Metric  model.Metric
	Summary *model.MetricSeriesSummary
	Points  []model.MetricSeriesPoint
}

func (s *Service) ListAnalyticsMetrics(ctx context.Context, p ListAnalyticsMetricsParams) ([]model.AnalyticsMetricSummary, error) {
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

	healthAllowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionHealthRead)
	if err != nil {
		return nil, err
	}

	source, validSource := normalizeAnalyticsSource(p.Source)
	if !validSource {
		return nil, ErrInvalidInput
	}
	if !healthAllowed {
		if source != nil && *source == "HEALTH" {
			return []model.AnalyticsMetricSummary{}, nil
		}
		userSource := "USER"
		source = &userSource
	}

	if p.DateFrom != nil && p.DateTo != nil && p.DateFrom.After(*p.DateTo) {
		return nil, ErrInvalidInput
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	items, err := s.repo.ListAnalyticsMetrics(ctx, repository.ListAnalyticsMetricsInput{
		PetID:    p.PetID,
		Q:        strings.TrimSpace(p.Q),
		DateFrom: p.DateFrom,
		DateTo:   p.DateTo,
		Source:   source,
		TypeIDs:  uniqueUUIDs(p.TypeIDs),
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) GetMetricSeries(ctx context.Context, p GetMetricSeriesParams) (*MetricSeriesResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.MetricID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	healthAllowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionHealthRead)
	if err != nil {
		return nil, err
	}

	source, validSource := normalizeAnalyticsSource(p.Source)
	if !validSource {
		return nil, ErrInvalidInput
	}
	if p.DateFrom != nil && p.DateTo != nil && p.DateFrom.After(*p.DateTo) {
		return nil, ErrInvalidInput
	}
	sort, validSort := normalizeMetricSeriesSort(p.Sort)
	if !validSort {
		return nil, ErrInvalidInput
	}

	metric, err := s.repo.GetMetricByID(ctx, p.PetID, p.MetricID)
	if err != nil {
		return nil, mapRepoErr(err)
	}

	if !healthAllowed {
		if source != nil && *source == "HEALTH" {
			summary := (*model.MetricSeriesSummary)(nil)
			if p.IncludeSummary {
				summary = &model.MetricSeriesSummary{}
			}
			return &MetricSeriesResult{
				Metric:  *metric,
				Summary: summary,
				Points:  []model.MetricSeriesPoint{},
			}, nil
		}
		userSource := "USER"
		source = &userSource
	}

	points, summary, err := s.repo.ListMetricSeries(ctx, repository.ListMetricSeriesInput{
		PetID:          p.PetID,
		MetricID:       p.MetricID,
		DateFrom:       p.DateFrom,
		DateTo:         p.DateTo,
		Source:         source,
		TypeIDs:        uniqueUUIDs(p.TypeIDs),
		Sort:           sort,
		IncludeSummary: p.IncludeSummary,
	})
	if err != nil {
		return nil, err
	}

	return &MetricSeriesResult{
		Metric:  *metric,
		Summary: summary,
		Points:  points,
	}, nil
}

func normalizeAnalyticsSource(raw string) (*string, bool) {
	v := strings.ToUpper(strings.TrimSpace(raw))
	if v == "" || v == "ALL" {
		return nil, true
	}
	if v == "USER" || v == "HEALTH" {
		return &v, true
	}
	return nil, false
}

func normalizeMetricSeriesSort(raw string) (string, bool) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return "occurred_at_asc", true
	}
	if v == "occurred_at_asc" || v == "occurred_at_desc" {
		return v, true
	}
	return "", false
}
