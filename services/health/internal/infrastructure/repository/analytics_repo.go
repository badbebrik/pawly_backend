package repository

import (
	"context"
	"fmt"
	"health/internal/model"
	repo "health/internal/repository"
	"strings"

	"github.com/google/uuid"
)

func (r *LogRepository) ListAnalyticsMetrics(ctx context.Context, in repo.ListAnalyticsMetricsInput) ([]model.AnalyticsMetricSummary, error) {
	if in.Limit <= 0 {
		in.Limit = 100
	}
	if in.Limit > 500 {
		in.Limit = 500
	}

	args := []any{in.PetID}
	pointWhere := []string{"l.pet_id = $1", "l.deleted_at IS NULL"}

	if in.Source != nil {
		args = append(args, *in.Source)
		pointWhere = append(pointWhere, fmt.Sprintf("l.source = $%d", len(args)))
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		pointWhere = append(pointWhere, fmt.Sprintf("l.occurred_at >= $%d", len(args)))
	}

	metricWhere := []string{"(m.scope = 'SYSTEM' OR m.pet_id = $1)", "m.deleted_at IS NULL"}
	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		metricWhere = append(metricWhere, fmt.Sprintf("m.name ILIKE $%d", len(args)))
	}

	args = append(args, in.Limit)
	limitIdx := len(args)

	query := fmt.Sprintf(`
		WITH point_rows AS (
			SELECT
				mv.metric_id,
				mv.value_num,
				l.occurred_at,
				l.id
			FROM metric_values mv
			JOIN logs l ON l.id = mv.log_id
			WHERE %s
		),
		agg AS (
			SELECT
				metric_id,
				COUNT(1)::int AS points_count,
				MIN(occurred_at) AS first_occurred_at,
				MAX(occurred_at) AS last_occurred_at
			FROM point_rows
			GROUP BY metric_id
		),
		last_point AS (
			SELECT DISTINCT ON (metric_id)
				metric_id,
				value_num
			FROM point_rows
			ORDER BY metric_id, occurred_at DESC, id DESC
		)
		SELECT
			m.id,
			m.name,
			m.scope,
			m.input_kind,
			m.unit_code,
			a.points_count,
			a.first_occurred_at,
			a.last_occurred_at,
			lp.value_num
		FROM agg a
		JOIN metrics m ON m.id = a.metric_id
		JOIN last_point lp ON lp.metric_id = a.metric_id
		WHERE %s
		ORDER BY a.last_occurred_at DESC, m.name ASC
		LIMIT $%d
	`, strings.Join(pointWhere, " AND "), strings.Join(metricWhere, " AND "), limitIdx)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.AnalyticsMetricSummary, 0, in.Limit)
	metricIDs := make([]uuid.UUID, 0, in.Limit)
	for rows.Next() {
		var item model.AnalyticsMetricSummary
		err := rows.Scan(
			&item.MetricID,
			&item.MetricName,
			&item.MetricScope,
			&item.InputKind,
			&item.UnitCode,
			&item.PointsCount,
			&item.FirstOccurredAt,
			&item.LastOccurredAt,
			&item.LastValueNum,
		)
		if err != nil {
			return nil, err
		}
		item.UsedInLogTypes = []model.AnalyticsUsedLogType{}
		items = append(items, item)
		metricIDs = append(metricIDs, item.MetricID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}

	usedMap, err := r.loadUsedLogTypes(ctx, in.PetID, metricIDs)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if used, ok := usedMap[items[i].MetricID]; ok {
			items[i].UsedInLogTypes = used
		}
	}

	return items, nil
}

func (r *LogRepository) loadUsedLogTypes(ctx context.Context, petID uuid.UUID, metricIDs []uuid.UUID) (map[uuid.UUID][]model.AnalyticsUsedLogType, error) {
	const query = `
		SELECT DISTINCT
			(req->>'metric_id')::uuid AS metric_id,
			lt.id,
			lt.name
		FROM log_types lt
		CROSS JOIN LATERAL jsonb_array_elements(lt.metric_requirements) req
		WHERE lt.deleted_at IS NULL
		  AND (lt.scope = 'SYSTEM' OR lt.pet_id = $1)
		  AND (req->>'metric_id')::uuid = ANY($2::uuid[])
		ORDER BY lt.name ASC
	`
	rows, err := r.db.Query(ctx, query, petID, metricIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]model.AnalyticsUsedLogType, len(metricIDs))
	for rows.Next() {
		var (
			metricID    uuid.UUID
			logTypeID   uuid.UUID
			logTypeName string
		)
		if err := rows.Scan(&metricID, &logTypeID, &logTypeName); err != nil {
			return nil, err
		}
		result[metricID] = append(result[metricID], model.AnalyticsUsedLogType{
			LogTypeID:   logTypeID,
			LogTypeName: logTypeName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

var _ repo.LogRepository = (*LogRepository)(nil)
