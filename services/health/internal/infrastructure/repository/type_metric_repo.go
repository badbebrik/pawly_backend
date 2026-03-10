package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"health/internal/model"
	repo "health/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type scanner interface {
	Scan(dest ...any) error
}

func (r *LogRepository) ListLogTypes(ctx context.Context, in repo.ListLogTypesInput) ([]model.LogType, error) {
	args := []any{in.PetID}
	where := []string{"(lt.scope = 'SYSTEM' OR lt.pet_id = $1)"}

	scope := strings.ToUpper(strings.TrimSpace(in.Scope))
	if scope == "SYSTEM" || scope == "CUSTOM" {
		args = append(args, scope)
		where = append(where, fmt.Sprintf("lt.scope = $%d", len(args)))
	}

	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf("lt.name ILIKE $%d", len(args)))
	}

	if !in.IncludeArchived {
		where = append(where, "lt.deleted_at IS NULL")
	}
	if in.OnlyWithMetrics {
		where = append(where, "jsonb_array_length(lt.metric_requirements) > 0")
	}

	query := fmt.Sprintf(`
		SELECT
			lt.id,
			lt.scope,
			lt.pet_id,
			lt.code,
			lt.name,
			lt.metric_requirements,
			lt.created_at,
			lt.created_by_user_id,
			lt.updated_at,
			lt.updated_by_user_id,
			lt.row_version,
			lt.deleted_at,
			lt.deleted_by_user_id
		FROM log_types lt
		WHERE %s
		ORDER BY lt.scope ASC, lt.name ASC
	`, strings.Join(where, " AND "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.LogType, 0)
	for rows.Next() {
		item, err := scanLogType(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) CreateLogType(ctx context.Context, in repo.CreateLogTypeInput) (*model.LogType, error) {
	reqBytes, err := json.Marshal(in.MetricRequirements)
	if err != nil {
		return nil, err
	}
	const query = `
		INSERT INTO log_types (
			id, scope, pet_id, code, name, metric_requirements,
			created_at, created_by_user_id,
			updated_at, updated_by_user_id,
			row_version, deleted_at, deleted_by_user_id
		)
		VALUES (
			$1, 'CUSTOM', $2, NULL, $3, $4,
			NOW(), $5,
			NOW(), $6,
			1, NULL, NULL
		)
	`
	_, err = r.db.Exec(ctx, query, in.ID, in.PetID, in.Name, reqBytes, in.CreatedByUserID, in.UpdatedByUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	return r.GetLogTypeByID(ctx, in.PetID, in.ID)
}

func (r *LogRepository) UpdateLogType(ctx context.Context, in repo.UpdateLogTypeInput) (*model.LogType, error) {
	reqBytes, err := json.Marshal(in.MetricRequirements)
	if err != nil {
		return nil, err
	}
	const query = `
		UPDATE log_types
		SET name = $4,
		    metric_requirements = $5,
		    updated_at = NOW(),
		    updated_by_user_id = $6,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND row_version = $3
		  AND scope = 'CUSTOM'
		  AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.Name, reqBytes, in.UpdatedByUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		exists, isCustom, err := r.lookupLogTypeState(ctx, in.PetID, in.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		if !isCustom {
			return nil, repo.ErrConflict
		}
		return nil, repo.ErrConflict
	}
	return r.GetLogTypeByID(ctx, in.PetID, in.ID)
}

func (r *LogRepository) ArchiveLogType(ctx context.Context, in repo.ArchiveLogTypeInput) error {
	const query = `
		UPDATE log_types
		SET deleted_at = NOW(),
		    deleted_by_user_id = $4,
		    updated_at = NOW(),
		    updated_by_user_id = $4,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND row_version = $3
		  AND scope = 'CUSTOM'
		  AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.DeletedByUserID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, isCustom, err := r.lookupLogTypeState(ctx, in.PetID, in.ID)
		if err != nil {
			return err
		}
		if !exists {
			return repo.ErrNotFound
		}
		if !isCustom {
			return repo.ErrConflict
		}
		return repo.ErrConflict
	}
	return nil
}

func (r *LogRepository) ListRecentLogTypes(ctx context.Context, petID uuid.UUID, includeHealth bool, limit int) ([]model.LogType, error) {
	if limit <= 0 {
		limit = 5
	}
	args := []any{petID, limit}
	where := []string{"l.pet_id = $1", "l.deleted_at IS NULL", "l.log_type_id IS NOT NULL"}
	if !includeHealth {
		where = append(where, "l.source = 'USER'")
	}

	query := fmt.Sprintf(`
		SELECT
			lt.id,
			lt.scope,
			lt.pet_id,
			lt.code,
			lt.name,
			lt.metric_requirements,
			lt.created_at,
			lt.created_by_user_id,
			lt.updated_at,
			lt.updated_by_user_id,
			lt.row_version,
			lt.deleted_at,
			lt.deleted_by_user_id
		FROM (
			SELECT l.log_type_id, MAX(l.occurred_at) AS last_occurred_at
			FROM logs l
			WHERE %s
			GROUP BY l.log_type_id
			ORDER BY MAX(l.occurred_at) DESC
			LIMIT $2
		) x
		JOIN log_types lt ON lt.id = x.log_type_id
		WHERE lt.deleted_at IS NULL
		ORDER BY x.last_occurred_at DESC
	`, strings.Join(where, " AND "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.LogType, 0, limit)
	for rows.Next() {
		item, err := scanLogType(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) ListMetrics(ctx context.Context, in repo.ListMetricsInput) ([]model.Metric, error) {
	args := []any{in.PetID}
	where := []string{"(m.scope = 'SYSTEM' OR m.pet_id = $1)"}

	scope := strings.ToUpper(strings.TrimSpace(in.Scope))
	if scope == "SYSTEM" || scope == "CUSTOM" {
		args = append(args, scope)
		where = append(where, fmt.Sprintf("m.scope = $%d", len(args)))
	}

	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf("m.name ILIKE $%d", len(args)))
	}

	if !in.IncludeArchived {
		where = append(where, "m.deleted_at IS NULL")
	}
	if in.OnlyWithData {
		where = append(where, `EXISTS (
			SELECT 1
			FROM metric_values mv
			JOIN logs l ON l.id = mv.log_id
			WHERE mv.metric_id = m.id
			  AND l.pet_id = $1
			  AND l.deleted_at IS NULL
		)`)
	}
	if in.OnlyWithUsage {
		where = append(where, `(
			EXISTS (
				SELECT 1
				FROM metric_values mv
				JOIN logs l ON l.id = mv.log_id
				WHERE mv.metric_id = m.id
				  AND l.pet_id = $1
				  AND l.deleted_at IS NULL
			)
			OR EXISTS (
				SELECT 1
				FROM log_types lt
				WHERE lt.deleted_at IS NULL
				  AND (lt.scope = 'SYSTEM' OR lt.pet_id = $1)
				  AND EXISTS (
					SELECT 1
					FROM jsonb_array_elements(lt.metric_requirements) req
					WHERE (req->>'metric_id')::uuid = m.id
				  )
			)
		)`)
	}

	query := fmt.Sprintf(`
		SELECT
			m.id,
			m.scope,
			m.pet_id,
			m.code,
			m.name,
			m.input_kind,
			m.unit_code,
			m.min_value,
			m.max_value,
			m.created_at,
			m.created_by_user_id,
			m.updated_at,
			m.updated_by_user_id,
			m.row_version,
			m.deleted_at,
			m.deleted_by_user_id,
			COALESCE((
				SELECT COUNT(1)::int
				FROM log_types lt
				WHERE lt.deleted_at IS NULL
				  AND (lt.scope = 'SYSTEM' OR lt.pet_id = $1)
				  AND EXISTS (
					SELECT 1
					FROM jsonb_array_elements(lt.metric_requirements) req
					WHERE (req->>'metric_id')::uuid = m.id
				  )
			), 0) AS log_types_count,
			COALESCE((
				SELECT COUNT(1)::int
				FROM metric_values mv
				JOIN logs l ON l.id = mv.log_id
				WHERE mv.metric_id = m.id
				  AND l.pet_id = $1
				  AND l.deleted_at IS NULL
			), 0) AS logs_count
		FROM metrics m
		WHERE %s
		ORDER BY m.scope ASC, m.name ASC
	`, strings.Join(where, " AND "))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Metric, 0)
	for rows.Next() {
		item, err := scanMetric(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) GetMetricByID(ctx context.Context, petID, metricID uuid.UUID) (*model.Metric, error) {
	const query = `
		SELECT
			m.id,
			m.scope,
			m.pet_id,
			m.code,
			m.name,
			m.input_kind,
			m.unit_code,
			m.min_value,
			m.max_value,
			m.created_at,
			m.created_by_user_id,
			m.updated_at,
			m.updated_by_user_id,
			m.row_version,
			m.deleted_at,
			m.deleted_by_user_id,
			COALESCE((
				SELECT COUNT(1)::int
				FROM log_types lt
				WHERE lt.deleted_at IS NULL
				  AND (lt.scope = 'SYSTEM' OR lt.pet_id = $2)
				  AND EXISTS (
					SELECT 1
					FROM jsonb_array_elements(lt.metric_requirements) req
					WHERE (req->>'metric_id')::uuid = m.id
				  )
			), 0) AS log_types_count,
			COALESCE((
				SELECT COUNT(1)::int
				FROM metric_values mv
				JOIN logs l ON l.id = mv.log_id
				WHERE mv.metric_id = m.id
				  AND l.pet_id = $2
				  AND l.deleted_at IS NULL
			), 0) AS logs_count
		FROM metrics m
		WHERE m.id = $1
		  AND (m.scope = 'SYSTEM' OR m.pet_id = $2)
	`
	item, err := scanMetric(r.db.QueryRow(ctx, query, metricID, petID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *LogRepository) CreateMetric(ctx context.Context, in repo.CreateMetricInput) (*model.Metric, error) {
	const query = `
		INSERT INTO metrics (
			id, scope, pet_id, code, name, input_kind, unit_code, min_value, max_value,
			created_at, created_by_user_id, updated_at, updated_by_user_id,
			row_version, deleted_at, deleted_by_user_id
		)
		VALUES (
			$1, 'CUSTOM', $2, NULL, $3, $4, $5, $6, $7,
			NOW(), $8, NOW(), $9,
			1, NULL, NULL
		)
	`
	_, err := r.db.Exec(ctx, query,
		in.ID,
		in.PetID,
		in.Name,
		in.InputKind,
		in.UnitCode,
		in.MinValue,
		in.MaxValue,
		in.CreatedByUserID,
		in.UpdatedByUserID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	return r.GetMetricByID(ctx, in.PetID, in.ID)
}

func (r *LogRepository) UpdateMetric(ctx context.Context, in repo.UpdateMetricInput) (*model.Metric, error) {
	const query = `
		UPDATE metrics
		SET name = $4,
		    input_kind = $5,
		    unit_code = $6,
		    min_value = $7,
		    max_value = $8,
		    updated_at = NOW(),
		    updated_by_user_id = $9,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND row_version = $3
		  AND scope = 'CUSTOM'
		  AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query,
		in.ID,
		in.PetID,
		in.RowVersion,
		in.Name,
		in.InputKind,
		in.UnitCode,
		in.MinValue,
		in.MaxValue,
		in.UpdatedByUserID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		exists, isCustom, err := r.lookupMetricState(ctx, in.PetID, in.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		if !isCustom {
			return nil, repo.ErrConflict
		}
		return nil, repo.ErrConflict
	}
	return r.GetMetricByID(ctx, in.PetID, in.ID)
}

func (r *LogRepository) ArchiveMetric(ctx context.Context, in repo.ArchiveMetricInput) error {
	const query = `
		UPDATE metrics
		SET deleted_at = NOW(),
		    deleted_by_user_id = $4,
		    updated_at = NOW(),
		    updated_by_user_id = $4,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND row_version = $3
		  AND scope = 'CUSTOM'
		  AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.DeletedByUserID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, isCustom, err := r.lookupMetricState(ctx, in.PetID, in.ID)
		if err != nil {
			return err
		}
		if !exists {
			return repo.ErrNotFound
		}
		if !isCustom {
			return repo.ErrConflict
		}
		return repo.ErrConflict
	}
	return nil
}

func (r *LogRepository) HasMetricValuesOutOfRange(ctx context.Context, petID, metricID uuid.UUID, minValue, maxValue *float64) (bool, error) {
	if minValue == nil && maxValue == nil {
		return false, nil
	}
	args := []any{petID, metricID}
	conds := make([]string, 0, 2)
	if minValue != nil {
		args = append(args, *minValue)
		conds = append(conds, fmt.Sprintf("mv.value_num < $%d", len(args)))
	}
	if maxValue != nil {
		args = append(args, *maxValue)
		conds = append(conds, fmt.Sprintf("mv.value_num > $%d", len(args)))
	}
	query := fmt.Sprintf(`
		SELECT 1
		FROM metric_values mv
		JOIN logs l ON l.id = mv.log_id
		WHERE l.pet_id = $1
		  AND l.deleted_at IS NULL
		  AND mv.metric_id = $2
		  AND (%s)
		LIMIT 1
	`, strings.Join(conds, " OR "))
	var x int
	err := r.db.QueryRow(ctx, query, args...).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func scanLogType(s scanner) (*model.LogType, error) {
	var (
		item     model.LogType
		reqBytes []byte
	)
	err := s.Scan(
		&item.ID,
		&item.Scope,
		&item.PetID,
		&item.Code,
		&item.Name,
		&reqBytes,
		&item.CreatedAt,
		&item.CreatedByUserID,
		&item.UpdatedAt,
		&item.UpdatedByUserID,
		&item.RowVersion,
		&item.DeletedAt,
		&item.DeletedByUserID,
	)
	if err != nil {
		return nil, err
	}
	if len(reqBytes) > 0 {
		if err := json.Unmarshal(reqBytes, &item.MetricRequirements); err != nil {
			return nil, err
		}
	}
	if item.MetricRequirements == nil {
		item.MetricRequirements = []model.LogTypeMetricRequirement{}
	}
	return &item, nil
}

func scanMetric(s scanner) (*model.Metric, error) {
	var item model.Metric
	err := s.Scan(
		&item.ID,
		&item.Scope,
		&item.PetID,
		&item.Code,
		&item.Name,
		&item.InputKind,
		&item.UnitCode,
		&item.MinValue,
		&item.MaxValue,
		&item.CreatedAt,
		&item.CreatedByUserID,
		&item.UpdatedAt,
		&item.UpdatedByUserID,
		&item.RowVersion,
		&item.DeletedAt,
		&item.DeletedByUserID,
		&item.Usage.LogTypesCount,
		&item.Usage.LogsCount,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *LogRepository) lookupLogTypeState(ctx context.Context, petID, logTypeID uuid.UUID) (bool, bool, error) {
	const query = `
		SELECT scope
		FROM log_types
		WHERE id = $1
		  AND (scope = 'SYSTEM' OR pet_id = $2)
	`
	var scope string
	err := r.db.QueryRow(ctx, query, logTypeID, petID).Scan(&scope)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, scope == "CUSTOM", nil
}

func (r *LogRepository) lookupMetricState(ctx context.Context, petID, metricID uuid.UUID) (bool, bool, error) {
	const query = `
		SELECT scope
		FROM metrics
		WHERE id = $1
		  AND (scope = 'SYSTEM' OR pet_id = $2)
	`
	var scope string
	err := r.db.QueryRow(ctx, query, metricID, petID).Scan(&scope)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, scope == "CUSTOM", nil
}
