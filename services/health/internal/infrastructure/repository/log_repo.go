package repository

import (
	"context"
	"errors"
	"fmt"
	"health/internal/model"
	repo "health/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LogRepository struct {
	db *pgxpool.Pool
}

func NewLogRepository(db *pgxpool.Pool) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) GetLog(ctx context.Context, petID, logID uuid.UUID) (*model.Log, error) {
	const query = `
		SELECT
			l.id,
			l.pet_id,
			l.occurred_at,
			l.log_type_id,
			lt.name,
			lt.scope,
			l.description,
			l.source,
			rel.right_entity_type,
			rel.right_entity_id,
			l.row_version,
			l.created_at,
			l.created_by_user_id,
			l.updated_at,
			l.updated_by_user_id,
			l.deleted_at,
			l.deleted_by_user_id
		FROM logs l
		LEFT JOIN log_types lt ON lt.id = l.log_type_id
		LEFT JOIN LATERAL (
			SELECT er.right_entity_type, er.right_entity_id
			FROM entity_relations er
			WHERE er.pet_id = l.pet_id
			  AND er.left_entity_type = 'LOG'
			  AND er.left_entity_id = l.id
			  AND er.right_entity_type IN ('PROCEDURE', 'VACCINATION')
			ORDER BY er.created_at ASC, er.id ASC
			LIMIT 1
		) rel ON TRUE
		WHERE l.id = $1 AND l.pet_id = $2 AND l.deleted_at IS NULL
	`

	logItem := model.Log{}
	err := r.db.QueryRow(ctx, query, logID, petID).Scan(
		&logItem.ID,
		&logItem.PetID,
		&logItem.OccurredAt,
		&logItem.LogTypeID,
		&logItem.LogTypeName,
		&logItem.LogTypeScope,
		&logItem.Description,
		&logItem.Source,
		&logItem.RelatedEntityType,
		&logItem.RelatedEntityID,
		&logItem.RowVersion,
		&logItem.CreatedAt,
		&logItem.CreatedByUserID,
		&logItem.UpdatedAt,
		&logItem.UpdatedByUserID,
		&logItem.DeletedAt,
		&logItem.DeletedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}

	metricValues, err := r.listMetricValues(ctx, logItem.ID)
	if err != nil {
		return nil, err
	}
	attachments, err := r.listAttachments(ctx, logItem.ID)
	if err != nil {
		return nil, err
	}

	logItem.MetricValues = metricValues
	logItem.Attachments = attachments
	return &logItem, nil
}

func (r *LogRepository) ListLogs(ctx context.Context, in repo.ListLogsInput) (repo.ListLogsOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 30
	}
	if in.Limit > 100 {
		in.Limit = 100
	}

	sort := strings.ToLower(strings.TrimSpace(in.Sort))
	if sort != "occurred_at_asc" {
		sort = "occurred_at_desc"
	}

	args := []any{in.PetID}
	where := []string{"l.pet_id = $1", "l.deleted_at IS NULL"}

	if in.Cursor != nil {
		args = append(args, in.Cursor.OccurredAt, in.Cursor.ID)
		if sort == "occurred_at_asc" {
			where = append(where, fmt.Sprintf("(l.occurred_at, l.id) > ($%d, $%d)", len(args)-1, len(args)))
		} else {
			where = append(where, fmt.Sprintf("(l.occurred_at, l.id) < ($%d, $%d)", len(args)-1, len(args)))
		}
	}

	if in.Q != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf("(COALESCE(lt.name, '') ILIKE $%d OR COALESCE(l.description, '') ILIKE $%d)", len(args), len(args)))
	}

	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("l.occurred_at >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("l.occurred_at <= $%d", len(args)))
	}
	if len(in.TypeIDs) > 0 {
		args = append(args, in.TypeIDs)
		where = append(where, fmt.Sprintf("l.log_type_id = ANY($%d::uuid[])", len(args)))
	}
	if in.Source != nil {
		args = append(args, *in.Source)
		where = append(where, fmt.Sprintf("l.source = $%d", len(args)))
	}
	if in.HasAttachments != nil {
		if *in.HasAttachments {
			where = append(where, "EXISTS (SELECT 1 FROM attachment_refs ar WHERE ar.entity_type = 'LOG' AND ar.entity_id = l.id)")
		} else {
			where = append(where, "NOT EXISTS (SELECT 1 FROM attachment_refs ar WHERE ar.entity_type = 'LOG' AND ar.entity_id = l.id)")
		}
	}
	if in.HasMetricValues != nil {
		if *in.HasMetricValues {
			where = append(where, "EXISTS (SELECT 1 FROM metric_values mv WHERE mv.log_id = l.id)")
		} else {
			where = append(where, "NOT EXISTS (SELECT 1 FROM metric_values mv WHERE mv.log_id = l.id)")
		}
	}

	orderDir := "DESC"
	if sort == "occurred_at_asc" {
		orderDir = "ASC"
	}

	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT
			l.id,
			l.pet_id,
			l.occurred_at,
			l.log_type_id,
			lt.name,
			lt.scope,
			l.description,
			l.source,
			rel.right_entity_type,
			rel.right_entity_id,
			l.row_version,
			l.created_by_user_id,
			COALESCE(att.cnt, 0) AS attachments_count
		FROM logs l
		LEFT JOIN log_types lt ON lt.id = l.log_type_id
		LEFT JOIN LATERAL (
			SELECT er.right_entity_type, er.right_entity_id
			FROM entity_relations er
			WHERE er.pet_id = l.pet_id
			  AND er.left_entity_type = 'LOG'
			  AND er.left_entity_id = l.id
			  AND er.right_entity_type IN ('PROCEDURE', 'VACCINATION')
			ORDER BY er.created_at ASC, er.id ASC
			LIMIT 1
		) rel ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt
			FROM attachment_refs ar
			WHERE ar.entity_type = 'LOG' AND ar.entity_id = l.id
		) att ON TRUE
		WHERE %s
		ORDER BY l.occurred_at %s, l.id %s
		LIMIT $%d
	`, strings.Join(where, " AND "), orderDir, orderDir, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListLogsOutput{}, err
	}
	defer rows.Close()

	items := make([]model.LogListItem, 0, in.Limit+1)
	for rows.Next() {
		var item model.LogListItem
		var description *string
		err := rows.Scan(
			&item.ID,
			&item.PetID,
			&item.OccurredAt,
			&item.LogTypeID,
			&item.LogTypeName,
			&item.LogTypeScope,
			&description,
			&item.Source,
			&item.RelatedEntityType,
			&item.RelatedEntityID,
			&item.RowVersion,
			&item.CreatedByUserID,
			&item.AttachmentsCount,
		)
		if err != nil {
			return repo.ListLogsOutput{}, err
		}
		item.DescriptionPreview = preview(description, 160)
		item.HasAttachments = item.AttachmentsCount > 0
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return repo.ListLogsOutput{}, err
	}

	out := repo.ListLogsOutput{Items: items}
	if len(items) > in.Limit {
		next := items[in.Limit]
		out.NextCursor = &repo.ListCursor{OccurredAt: next.OccurredAt, ID: next.ID}
		out.Items = items[:in.Limit]
	}

	return out, nil
}

func (r *LogRepository) CreateLog(ctx context.Context, in repo.CreateLogInput) (*model.Log, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insertQuery = `
		INSERT INTO logs (
			id,
			pet_id,
			occurred_at,
			log_type_id,
			description,
			source,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			1,
			NOW(),
			$7,
			NOW(),
			$8
		)
	`
	_, err = tx.Exec(ctx, insertQuery,
		in.ID,
		in.PetID,
		in.OccurredAt,
		in.LogTypeID,
		in.Description,
		in.Source,
		in.CreatedByUserID,
		in.UpdatedByUserID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if err := r.replaceMetricValuesTx(ctx, tx, in.ID, in.MetricValues); err != nil {
		return nil, err
	}
	if err := r.replaceAttachmentsTx(ctx, tx, in.PetID, in.ID, in.AttachmentFileIDs, in.CreatedByUserID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetLog(ctx, in.PetID, in.ID)
}

func (r *LogRepository) UpdateLog(ctx context.Context, in repo.UpdateLogInput) (*model.Log, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const updateQuery = `
		UPDATE logs
		SET occurred_at = $4,
		    log_type_id = $5,
		    description = $6,
		    updated_at = NOW(),
		    updated_by_user_id = $7,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND row_version = $3
		  AND source = 'USER'
		  AND deleted_at IS NULL
	`

	cmd, err := tx.Exec(ctx, updateQuery,
		in.ID,
		in.PetID,
		in.RowVersion,
		in.OccurredAt,
		in.LogTypeID,
		in.Description,
		in.UpdatedByUserID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if cmd.RowsAffected() == 0 {
		exists, source, err := r.lookupLogStateTx(ctx, tx, in.PetID, in.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		if source != "USER" {
			return nil, repo.ErrConflict
		}
		return nil, repo.ErrConflict
	}

	if err := r.replaceMetricValuesTx(ctx, tx, in.ID, in.MetricValues); err != nil {
		return nil, err
	}
	if err := r.replaceAttachmentsTx(ctx, tx, in.PetID, in.ID, in.AttachmentFileIDs, in.UpdatedByUserID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetLog(ctx, in.PetID, in.ID)
}

func (r *LogRepository) SoftDeleteLog(ctx context.Context, in repo.DeleteLogInput) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
		UPDATE logs
		SET deleted_at = NOW(),
		    deleted_by_user_id = $4,
		    updated_at = NOW(),
		    updated_by_user_id = $4,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND row_version = $3
		  AND source = 'USER'
		  AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.DeletedByUserID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, source, err := r.lookupLogStateTx(ctx, tx, in.PetID, in.ID)
		if err != nil {
			return err
		}
		if !exists {
			return repo.ErrNotFound
		}
		if source != "USER" {
			return repo.ErrConflict
		}
		return repo.ErrConflict
	}
	if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = 'LOG' AND entity_id = $1`, in.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM entity_relations
		WHERE pet_id = $1
		  AND (
			(left_entity_type = 'LOG' AND left_entity_id = $2)
			OR
			(right_entity_type = 'LOG' AND right_entity_id = $2)
		  )
	`, in.PetID, in.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LogRepository) UpsertHealthEntityLog(ctx context.Context, in repo.UpsertHealthEntityLogInput) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	logID, err := r.lookupHealthEntityLogIDTx(ctx, tx, in.PetID, in.EntityType, in.EntityID)
	if err != nil {
		return err
	}
	if logID != nil {
		_, err := tx.Exec(ctx, `
			UPDATE logs
			SET occurred_at = $3,
			    description = $4,
			    updated_at = NOW(),
			    updated_by_user_id = $5,
			    row_version = row_version + 1
			WHERE id = $1
			  AND pet_id = $2
			  AND source = 'HEALTH'
			  AND deleted_at IS NULL
		`, *logID, in.PetID, in.OccurredAt, in.Description, in.UpdatedByUserID)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	newLogID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO logs (
			id, pet_id, occurred_at, log_type_id, description, source, row_version,
			created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES (
			$1, $2, $3, NULL, $4, 'HEALTH', 1,
			NOW(), $5, NOW(), $6
		)
	`, newLogID, in.PetID, in.OccurredAt, in.Description, in.CreatedByUserID, in.UpdatedByUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrConflict
		}
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO entity_relations (
			id, pet_id, left_entity_type, left_entity_id, right_entity_type, right_entity_id, created_by_user_id, created_at
		) VALUES ($1, $2, 'LOG', $3, $4, $5, $6, NOW())
	`, uuid.New(), in.PetID, newLogID, in.EntityType, in.EntityID, in.CreatedByUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrConflict
		}
		return err
	}
	return tx.Commit(ctx)
}

func (r *LogRepository) DeleteHealthEntityLog(ctx context.Context, in repo.DeleteHealthEntityLogInput) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	logID, err := r.lookupHealthEntityLogIDTx(ctx, tx, in.PetID, in.EntityType, in.EntityID)
	if err != nil {
		return err
	}
	if logID == nil {
		return tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `
		UPDATE logs
		SET deleted_at = NOW(),
		    deleted_by_user_id = $3,
		    updated_at = NOW(),
		    updated_by_user_id = $3,
		    row_version = row_version + 1
		WHERE id = $1
		  AND pet_id = $2
		  AND source = 'HEALTH'
		  AND deleted_at IS NULL
	`, *logID, in.PetID, in.DeletedByUserID)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = 'LOG' AND entity_id = $1`, *logID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM entity_relations
		WHERE pet_id = $1
		  AND (
			(left_entity_type = 'LOG' AND left_entity_id = $2)
			OR
			(right_entity_type = 'LOG' AND right_entity_id = $2)
		  )
	`, in.PetID, *logID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LogRepository) GetLogTypeByID(ctx context.Context, petID uuid.UUID, logTypeID uuid.UUID) (*model.LogType, error) {
	const query = `
		SELECT
			id,
			scope,
			pet_id,
			code,
			name,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id,
			row_version,
			deleted_at,
			deleted_by_user_id
		FROM log_types
		WHERE id = $1
		  AND (scope = 'SYSTEM' OR pet_id = $2)
		  AND deleted_at IS NULL
	`
	item, err := scanLogTypeBase(r.db.QueryRow(ctx, query, logTypeID, petID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	reqs, err := r.loadLogTypeRequirements(ctx, []uuid.UUID{item.ID})
	if err != nil {
		return nil, err
	}
	item.MetricRequirements = reqs[item.ID]
	return item, nil
}

func (r *LogRepository) GetMetricsByIDs(ctx context.Context, petID uuid.UUID, metricIDs []uuid.UUID) (map[uuid.UUID]model.Metric, error) {
	if len(metricIDs) == 0 {
		return map[uuid.UUID]model.Metric{}, nil
	}
	const query = `
			SELECT id, scope, pet_id, code, name, input_kind, unit, min_value, max_value
			FROM metrics
		WHERE id = ANY($1::uuid[])
		  AND deleted_at IS NULL
		  AND (scope = 'SYSTEM' OR pet_id = $2)
	`
	rows, err := r.db.Query(ctx, query, metricIDs, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]model.Metric, len(metricIDs))
	for rows.Next() {
		var item model.Metric
		err := rows.Scan(
			&item.ID,
			&item.Scope,
			&item.PetID,
			&item.Code,
			&item.Name,
			&item.InputKind,
			&item.Unit,
			&item.MinValue,
			&item.MaxValue,
		)
		if err != nil {
			return nil, err
		}
		result[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *LogRepository) replaceMetricValuesTx(ctx context.Context, tx pgx.Tx, logID uuid.UUID, values []repo.LogMetricValueInput) error {
	if _, err := tx.Exec(ctx, `DELETE FROM metric_values WHERE log_id = $1`, logID); err != nil {
		return err
	}
	for i := range values {
		const query = `
			INSERT INTO metric_values (id, log_id, metric_id, value_num, created_at)
			VALUES ($1, $2, $3, $4, NOW())
		`
		_, err := tx.Exec(ctx, query, uuid.New(), logID, values[i].MetricID, values[i].ValueNum)
		if err != nil {
			if isUniqueViolation(err) {
				return repo.ErrConflict
			}
			return err
		}
	}
	return nil
}

func (r *LogRepository) replaceAttachmentsTx(ctx context.Context, tx pgx.Tx, petID, logID uuid.UUID, fileIDs []uuid.UUID, addedBy uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = 'LOG' AND entity_id = $1`, logID); err != nil {
		return err
	}
	for i := range fileIDs {
		const query = `
			INSERT INTO attachment_refs (id, pet_id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at)
			VALUES ($1, $2, 'LOG', $3, $4, NULL, 'other', $5, NOW())
		`
		_, err := tx.Exec(ctx, query, uuid.New(), petID, logID, fileIDs[i], addedBy)
		if err != nil {
			if isUniqueViolation(err) {
				return repo.ErrConflict
			}
			return err
		}
	}
	return nil
}

func (r *LogRepository) listMetricValues(ctx context.Context, logID uuid.UUID) ([]model.LogMetricValue, error) {
	const query = `
		SELECT
			mv.metric_id,
			m.name,
			m.input_kind,
			m.unit,
			mv.value_num
		FROM metric_values mv
		JOIN metrics m ON m.id = mv.metric_id
		WHERE mv.log_id = $1
		ORDER BY mv.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, logID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.LogMetricValue, 0)
	for rows.Next() {
		var item model.LogMetricValue
		err := rows.Scan(
			&item.MetricID,
			&item.MetricName,
			&item.InputKind,
			&item.Unit,
			&item.ValueNum,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) listAttachments(ctx context.Context, logID uuid.UUID) ([]model.LogAttachment, error) {
	const query = `
		SELECT
			id,
			file_id,
			file_name,
			file_type,
			added_by_user_id,
			added_at
		FROM attachment_refs
		WHERE entity_type = 'LOG' AND entity_id = $1
		ORDER BY added_at ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query, logID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.LogAttachment, 0)
	for rows.Next() {
		var item model.LogAttachment
		err := rows.Scan(
			&item.ID,
			&item.FileID,
			&item.FileName,
			&item.FileType,
			&item.AddedByUserID,
			&item.AddedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) lookupLogState(ctx context.Context, petID, logID uuid.UUID) (bool, string, error) {
	const query = `
		SELECT source
		FROM logs
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var source string
	err := r.db.QueryRow(ctx, query, logID, petID).Scan(&source)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, source, nil
}

func (r *LogRepository) lookupLogStateTx(ctx context.Context, tx pgx.Tx, petID, logID uuid.UUID) (bool, string, error) {
	const query = `
		SELECT source
		FROM logs
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var source string
	err := tx.QueryRow(ctx, query, logID, petID).Scan(&source)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, source, nil
}

func (r *LogRepository) lookupHealthEntityLogIDTx(ctx context.Context, tx pgx.Tx, petID uuid.UUID, entityType string, entityID uuid.UUID) (*uuid.UUID, error) {
	const query = `
		SELECT l.id
		FROM entity_relations rel
		JOIN logs l ON l.id = rel.left_entity_id
		WHERE rel.pet_id = $1
		  AND rel.left_entity_type = 'LOG'
		  AND rel.right_entity_type = $2
		  AND rel.right_entity_id = $3
		  AND l.pet_id = $1
		  AND l.source = 'HEALTH'
		  AND l.deleted_at IS NULL
		ORDER BY l.created_at ASC, l.id ASC
		LIMIT 1
	`
	var logID uuid.UUID
	err := tx.QueryRow(ctx, query, petID, entityType, entityID).Scan(&logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &logID, nil
}

func preview(v *string, limit int) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	r := []rune(trimmed)
	if len(r) <= limit {
		return &trimmed
	}
	s := string(r[:limit])
	return &s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var _ repo.LogRepository = (*LogRepository)(nil)
