package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Analytics struct {
	repo ports.LogsRepository
	acl  ports.HealthAccessChecker
}

func NewAnalytics(repo ports.LogsRepository, acl ports.HealthAccessChecker) *Analytics {
	return &Analytics{repo: repo, acl: acl}
}

type ListAnalyticsMetricsParams struct {
	UserID   uuid.UUID
	PetID    uuid.UUID
	Q        string
	DateFrom *time.Time
	DateTo   *time.Time
	TypeIDs  []uuid.UUID
	Limit    int
}

type GetMetricSeriesParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	MetricID       uuid.UUID
	DateFrom       *time.Time
	DateTo         *time.Time
	TypeIDs        []uuid.UUID
	Sort           string
	IncludeSummary bool
}

type MetricSeriesResult struct {
	Metric  model.Metric
	Summary *model.MetricSeriesSummary
	Points  []model.MetricSeriesPoint
}

func (u *Analytics) ListAnalyticsMetrics(ctx context.Context, p ListAnalyticsMetricsParams) ([]model.AnalyticsMetricSummary, error) {
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
	return u.repo.ListAnalyticsMetrics(ctx, ports.ListAnalyticsMetricsInput{
		PetID:    p.PetID,
		Q:        strings.TrimSpace(p.Q),
		DateFrom: p.DateFrom,
		DateTo:   p.DateTo,
		TypeIDs:  uniqueUUIDs(p.TypeIDs),
		Limit:    limit,
	})
}

func (u *Analytics) GetMetricSeries(ctx context.Context, p GetMetricSeriesParams) (*MetricSeriesResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.MetricID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	if p.DateFrom != nil && p.DateTo != nil && p.DateFrom.After(*p.DateTo) {
		return nil, ErrInvalidInput
	}
	sort, valid := normalizeMetricSeriesSort(p.Sort)
	if !valid {
		return nil, ErrInvalidInput
	}
	metric, err := u.repo.GetMetricByID(ctx, p.PetID, p.MetricID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	points, summary, err := u.repo.ListMetricSeries(ctx, ports.ListMetricSeriesInput{
		PetID:          p.PetID,
		MetricID:       p.MetricID,
		DateFrom:       p.DateFrom,
		DateTo:         p.DateTo,
		TypeIDs:        uniqueUUIDs(p.TypeIDs),
		Sort:           sort,
		IncludeSummary: p.IncludeSummary,
	})
	if err != nil {
		return nil, err
	}
	return &MetricSeriesResult{Metric: *metric, Summary: summary, Points: points}, nil
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
