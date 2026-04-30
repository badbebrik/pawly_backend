package repository

import (
	"context"
	"errors"
	"fmt"
	ports "health/internal/application/ports"
	"health/internal/domain/model"
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
			return nil, ports.ErrNotFound
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

func (r *LogRepository) ListLogs(ctx context.Context, in ports.ListLogsQuery) (ports.ListLogsResult, error) {
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
		return ports.ListLogsResult{}, err
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
			&item.RelatedEntityType,
			&item.RelatedEntityID,
			&item.RowVersion,
			&item.CreatedByUserID,
			&item.AttachmentsCount,
		)
		if err != nil {
			return ports.ListLogsResult{}, err
		}
		item.DescriptionPreview = preview(description, 160)
		item.HasAttachments = item.AttachmentsCount > 0
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ports.ListLogsResult{}, err
	}

	out := ports.ListLogsResult{Items: items}
	if len(items) > in.Limit {
		next := items[in.Limit]
		out.NextCursor = &ports.LogCursor{OccurredAt: next.OccurredAt, ID: next.ID}
		out.Items = items[:in.Limit]
	}
	if err := r.attachMetricValuesToLogList(ctx, out.Items); err != nil {
		return ports.ListLogsResult{}, err
	}

	return out, nil
}

func (r *LogRepository) CreateLog(ctx context.Context, in ports.CreateLogInput) (*model.Log, error) {
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
			1,
			NOW(),
			$6,
			NOW(),
			$7
		)
	`
	_, err = tx.Exec(ctx, insertQuery,
		in.ID,
		in.PetID,
		in.OccurredAt,
		in.LogTypeID,
		in.Description,
		in.CreatedByUserID,
		in.UpdatedByUserID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrConflict
		}
		return nil, err
	}

	if err := r.replaceMetricValuesTx(ctx, tx, in.ID, in.MetricValues); err != nil {
		return nil, err
	}
	if err := r.replaceAttachmentsTx(ctx, tx, in.PetID, in.ID, in.Attachments, in.CreatedByUserID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetLog(ctx, in.PetID, in.ID)
}

func (r *LogRepository) UpdateLog(ctx context.Context, in ports.UpdateLogInput) (*model.Log, error) {
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
			return nil, ports.ErrConflict
		}
		return nil, err
	}

	if cmd.RowsAffected() == 0 {
		exists, err := r.logExistsTx(ctx, tx, in.PetID, in.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ports.ErrNotFound
		}
		return nil, ports.ErrConflict
	}

	if err := r.replaceMetricValuesTx(ctx, tx, in.ID, in.MetricValues); err != nil {
		return nil, err
	}
	if err := r.replaceAttachmentsTx(ctx, tx, in.PetID, in.ID, in.Attachments, in.UpdatedByUserID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetLog(ctx, in.PetID, in.ID)
}

func (r *LogRepository) SoftDeleteLog(ctx context.Context, in ports.DeleteLogInput) error {
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
		  AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.DeletedByUserID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, err := r.logExistsTx(ctx, tx, in.PetID, in.ID)
		if err != nil {
			return err
		}
		if !exists {
			return ports.ErrNotFound
		}
		return ports.ErrConflict
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
			return nil, ports.ErrNotFound
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

func (r *LogRepository) replaceMetricValuesTx(ctx context.Context, tx pgx.Tx, logID uuid.UUID, values []ports.LogMetricValueInput) error {
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
				return ports.ErrConflict
			}
			return err
		}
	}
	return nil
}

func (r *LogRepository) replaceAttachmentsTx(ctx context.Context, tx pgx.Tx, petID, logID uuid.UUID, attachments []ports.AttachmentInput, addedBy uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = 'LOG' AND entity_id = $1`, logID); err != nil {
		return err
	}
	for i := range attachments {
		att := attachments[i]
		const query = `
			INSERT INTO attachment_refs (id, pet_id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at)
			VALUES ($1, $2, 'LOG', $3, $4, $5, $6, $7, NOW())
		`
		_, err := tx.Exec(ctx, query, uuid.New(), petID, logID, att.FileID, att.FileName, att.FileType, addedBy)
		if err != nil {
			if isUniqueViolation(err) {
				return ports.ErrConflict
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

func (r *LogRepository) attachMetricValuesToLogList(ctx context.Context, items []model.LogListItem) error {
	if len(items) == 0 {
		return nil
	}
	logIDs := make([]uuid.UUID, 0, len(items))
	indexByLogID := make(map[uuid.UUID]int, len(items))
	for i := range items {
		logIDs = append(logIDs, items[i].ID)
		indexByLogID[items[i].ID] = i
	}

	const query = `
		SELECT
			mv.log_id,
			mv.metric_id,
			m.name,
			m.input_kind,
			m.unit,
			mv.value_num
		FROM metric_values mv
		JOIN metrics m ON m.id = mv.metric_id
		WHERE mv.log_id = ANY($1::uuid[])
		ORDER BY mv.log_id ASC, mv.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, logIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var logID uuid.UUID
		var value model.LogMetricValue
		if err := rows.Scan(
			&logID,
			&value.MetricID,
			&value.MetricName,
			&value.InputKind,
			&value.Unit,
			&value.ValueNum,
		); err != nil {
			return err
		}
		idx, ok := indexByLogID[logID]
		if !ok {
			continue
		}
		items[idx].MetricValues = append(items[idx].MetricValues, value)
	}
	return rows.Err()
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

func (r *LogRepository) logExistsTx(ctx context.Context, tx pgx.Tx, petID, logID uuid.UUID) (bool, error) {
	const query = `
		SELECT 1
		FROM logs
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var exists int
	err := tx.QueryRow(ctx, query, logID, petID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

var _ ports.LogsRepository = (*LogRepository)(nil)
