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
	UserID uuid.UUID
	PetID  uuid.UUID
	Q      string
	Range  string
	Source string
	Limit  int
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

	dateFrom, ok := normalizeRangeDateFrom(p.Range)
	if !ok {
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
		DateFrom: dateFrom,
		Source:   source,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	return items, nil
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

func normalizeRangeDateFrom(raw string) (*time.Time, bool) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" || v == "all" {
		return nil, true
	}
	var days int
	switch v {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		return nil, false
	}
	t := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	return &t, true
}
