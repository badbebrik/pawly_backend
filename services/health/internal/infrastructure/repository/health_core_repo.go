package repository

import (
	"context"
	"errors"
	"fmt"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *LogRepository) ListHealthDictionaryItems(ctx context.Context, in ports.ListHealthDictionaryItemsInput) ([]model.HealthDictionaryItem, error) {
	args := []any{in.PetID}
	where := []string{"is_archived = FALSE", "(pet_id IS NULL OR pet_id = $1)"}
	if len(in.Kinds) > 0 {
		args = append(args, in.Kinds)
		where = append(where, fmt.Sprintf("kind = ANY($%d)", len(args)))
	}
	query := fmt.Sprintf(`
		SELECT id, kind, pet_id, code, name, is_system, is_archived, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM health_dictionary_items
		WHERE %s
		ORDER BY kind ASC, is_system DESC, name ASC, id ASC
	`, strings.Join(where, " AND "))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.HealthDictionaryItem, 0)
	for rows.Next() {
		item, err := scanHealthDictionaryItem(rows)
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

func (r *LogRepository) GetHealthDictionaryItem(ctx context.Context, petID, itemID uuid.UUID, kind string) (*model.HealthDictionaryItem, error) {
	const query = `
		SELECT id, kind, pet_id, code, name, is_system, is_archived, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM health_dictionary_items
		WHERE id = $1 AND kind = $2 AND (pet_id IS NULL OR pet_id = $3)
	`
	item, err := scanHealthDictionaryItemRow(r.db.QueryRow(ctx, query, itemID, kind, petID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *LogRepository) GetOrCreateCustomHealthDictionaryItem(ctx context.Context, in ports.GetOrCreateCustomHealthDictionaryItemInput) (*model.HealthDictionaryItem, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ports.ErrConflict
	}
	const selectQuery = `
		SELECT id, kind, pet_id, code, name, is_system, is_archived, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM health_dictionary_items
		WHERE kind = $1 AND pet_id = $2 AND lower(name) = lower($3) AND is_system = FALSE AND is_archived = FALSE
	`
	item, err := scanHealthDictionaryItemRow(r.db.QueryRow(ctx, selectQuery, in.Kind, in.PetID, name))
	if err == nil {
		return &item, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	id := uuid.New()
	const insertQuery = `
		INSERT INTO health_dictionary_items (
			id, kind, pet_id, code, name, is_system, is_archived, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,NULL,$4,FALSE,FALSE,NOW(),$5,NOW(),$6)
	`
	if _, err := r.db.Exec(ctx, insertQuery, id, in.Kind, in.PetID, name, in.CreatedBy, in.UpdatedBy); err != nil {
		if !isUniqueViolation(err) {
			return nil, err
		}
		item, err := scanHealthDictionaryItemRow(r.db.QueryRow(ctx, selectQuery, in.Kind, in.PetID, name))
		if err != nil {
			return nil, err
		}
		return &item, nil
	}
	return r.GetHealthDictionaryItem(ctx, in.PetID, id, in.Kind)
}

func (r *LogRepository) GetVetVisit(ctx context.Context, petID, visitID uuid.UUID, includeRelatedLogs bool) (*model.VetVisit, error) {
	const query = `
		SELECT
			id,
			pet_id,
			status,
			visit_type,
			title,
			scheduled_at,
			completed_at,
			reason_text,
			result_text,
			clinic_name,
			vet_name,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id
		FROM vet_visits
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.VetVisit
	err := r.db.QueryRow(ctx, query, visitID, petID).Scan(
		&item.ID,
		&item.PetID,
		&item.Status,
		&item.VisitType,
		&item.Title,
		&item.ScheduledAt,
		&item.CompletedAt,
		&item.ReasonText,
		&item.ResultText,
		&item.ClinicName,
		&item.VetName,
		&item.RowVersion,
		&item.CreatedAt,
		&item.CreatedByUserID,
		&item.UpdatedAt,
		&item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	attachments, err := r.listHealthAttachments(ctx, "VET_VISIT", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	item.RelatedLogs = []model.RelatedLog{}
	if includeRelatedLogs {
		relatedLogs, err := r.listVetVisitRelatedLogs(ctx, item.PetID, item.ID)
		if err != nil {
			return nil, err
		}
		item.RelatedLogs = relatedLogs
	}
	return &item, nil
}

func (r *LogRepository) ListVetVisits(ctx context.Context, in ports.ListVetVisitsQuery) (ports.ListVetVisitsResult, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeVetVisitSort(in.Sort)
	if !ok {
		return ports.ListVetVisitsResult{}, ports.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"vv.pet_id = $1", "vv.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("vv.status = $%d", len(args)))
	}
	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf(`(
			vv.visit_type ILIKE $%d
			OR COALESCE(vv.title, '') ILIKE $%d
			OR COALESCE(vv.reason_text, '') ILIKE $%d
			OR COALESCE(vv.result_text, '') ILIKE $%d
			OR COALESCE(vv.clinic_name, '') ILIKE $%d
			OR COALESCE(vv.vet_name, '') ILIKE $%d
		)`, len(args), len(args), len(args), len(args), len(args), len(args)))
	}
	switch in.Bucket {
	case "upcoming":
		where = append(where, "vv.status = 'PLANNED'")
	case "history":
		where = append(where, "vv.status = 'COMPLETED'")
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("vv.scheduled_at >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("vv.scheduled_at <= $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(%s, vv.id) %s ($%d, $%d)", orderExpr, cursorOp, len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT
			vv.id,
			vv.pet_id,
			vv.status,
			vv.visit_type,
			vv.title,
			vv.scheduled_at,
			vv.completed_at,
			vv.reason_text,
			vv.result_text,
			vv.clinic_name,
			vv.vet_name,
			COALESCE(rl.cnt, 0) AS related_logs_count,
			COALESCE(att.cnt, 0) AS attachments_count,
			vv.row_version,
			vv.created_at,
			vv.created_by_user_id,
			vv.updated_at,
			vv.updated_by_user_id,
			%s AS cursor_sort_at
		FROM vet_visits vv
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt
			FROM entity_relations rel
			WHERE rel.pet_id = vv.pet_id
			  AND rel.left_entity_type = 'LOG'
			  AND rel.right_entity_type = 'VET_VISIT'
			  AND rel.right_entity_id = vv.id
		) rl ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt
			FROM attachment_refs ar
			WHERE ar.entity_type = 'VET_VISIT' AND ar.entity_id = vv.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, vv.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ports.ListVetVisitsResult{}, err
	}
	defer rows.Close()

	items := make([]model.VetVisitListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.VetVisitListItem
		var cursorTime time.Time
		if err := rows.Scan(
			&item.ID,
			&item.PetID,
			&item.Status,
			&item.VisitType,
			&item.Title,
			&item.ScheduledAt,
			&item.CompletedAt,
			&item.ReasonText,
			&item.ResultText,
			&item.ClinicName,
			&item.VetName,
			&item.RelatedLogsCount,
			&item.AttachmentsCount,
			&item.RowVersion,
			&item.CreatedAt,
			&item.CreatedByUserID,
			&item.UpdatedAt,
			&item.UpdatedByUserID,
			&cursorTime,
		); err != nil {
			return ports.ListVetVisitsResult{}, err
		}
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return ports.ListVetVisitsResult{}, err
	}

	out := ports.ListVetVisitsResult{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &ports.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateVetVisit(ctx context.Context, in ports.CreateVetVisitInput) (*model.VetVisit, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO vet_visits (
			id, pet_id, status, visit_type, title, scheduled_at, completed_at, reason_text, result_text, clinic_name, vet_name,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1,NOW(),$12,NOW(),$13)
	`
	_, err = tx.Exec(ctx, query,
		in.ID, in.PetID, in.Status, in.VisitType, in.Title, in.ScheduledAt, in.CompletedAt, in.ReasonText, in.ResultText, in.ClinicName, in.VetName,
		in.CreatedBy, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VET_VISIT", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetVetVisit(ctx, in.PetID, in.ID, true)
	return item, sync, err
}

func (r *LogRepository) UpdateVetVisit(ctx context.Context, in ports.UpdateVetVisitInput) (*model.VetVisit, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE vet_visits
		SET status = $4,
		    visit_type = $5,
		    title = $6,
		    scheduled_at = $7,
		    completed_at = $8,
		    reason_text = $9,
		    result_text = $10,
		    clinic_name = $11,
		    vet_name = $12,
		    updated_at = NOW(),
		    updated_by_user_id = $13,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query,
		in.ID, in.PetID, in.RowVersion, in.Status, in.VisitType, in.Title, in.ScheduledAt, in.CompletedAt, in.ReasonText, in.ResultText, in.ClinicName, in.VetName, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "vet_visits", in.ID, in.PetID); err != nil {
			return nil, ports.AttachmentSync{}, err
		} else if !exists {
			return nil, ports.AttachmentSync{}, ports.ErrNotFound
		}
		return nil, ports.AttachmentSync{}, ports.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VET_VISIT", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetVetVisit(ctx, in.PetID, in.ID, true)
	return item, sync, err
}

func (r *LogRepository) DeleteVetVisit(ctx context.Context, in ports.DeleteVetVisitInput) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE vet_visits
		SET deleted_at = NOW(),
		    deleted_by_user_id = $4,
		    updated_at = NOW(),
		    updated_by_user_id = $4,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.DeletedBy)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "vet_visits", in.ID, in.PetID); err != nil {
			return err
		} else if !exists {
			return ports.ErrNotFound
		}
		return ports.ErrConflict
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM entity_relations
		WHERE pet_id = $1
		  AND right_entity_type = 'VET_VISIT'
		  AND right_entity_id = $2
	`, in.PetID, in.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = 'VET_VISIT' AND entity_id = $1`, in.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LogRepository) LinkVetVisitLog(ctx context.Context, petID, visitID, logID, addedBy uuid.UUID) (*model.RelatedLog, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if exists, err := r.healthEntityExistsTx(ctx, tx, "vet_visits", visitID, petID); err != nil {
		return nil, err
	} else if !exists {
		return nil, ports.ErrNotFound
	}
	if exists, err := r.logExistsTx(ctx, tx, petID, logID); err != nil {
		return nil, err
	} else if !exists {
		return nil, ports.ErrNotFound
	}
	const query = `
		INSERT INTO entity_relations (
			id, pet_id, left_entity_type, left_entity_id, right_entity_type, right_entity_id, created_by_user_id, created_at
	) VALUES ($1,$2,'LOG',$3,'VET_VISIT',$4,$5,NOW())
	`
	if _, err := tx.Exec(ctx, query, uuid.New(), petID, logID, visitID, addedBy); err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrConflict
		}
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	item, err := r.getRelatedLog(ctx, petID, logID)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *LogRepository) UnlinkVetVisitLog(ctx context.Context, petID, visitID, logID uuid.UUID) error {
	const query = `
		DELETE FROM entity_relations rel
		WHERE rel.pet_id = $1
		  AND rel.left_entity_type = 'LOG'
		  AND rel.left_entity_id = $2
		  AND rel.right_entity_type = 'VET_VISIT'
		  AND rel.right_entity_id = $3
	`
	cmd, err := r.db.Exec(ctx, query, petID, logID, visitID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *LogRepository) GetVaccination(ctx context.Context, petID, vaccinationID uuid.UUID) (*model.Vaccination, error) {
	const query = `
		SELECT id, pet_id, generated_from_id, status, vaccine_name, catalog_medication_id, scheduled_at, administered_at, next_due_at,
		       (
				SELECT rel.right_entity_id
				FROM entity_relations rel
				JOIN vet_visits vv ON vv.id = rel.right_entity_id AND vv.pet_id = rel.pet_id AND vv.deleted_at IS NULL
				WHERE rel.pet_id = v.pet_id
				  AND rel.left_entity_type = 'VACCINATION'
				  AND rel.left_entity_id = v.id
				  AND rel.right_entity_type = 'VET_VISIT'
				LIMIT 1
		       ) AS vet_visit_id,
		       clinic_name, vet_name, notes, row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM vaccinations v
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.Vaccination
	err := r.db.QueryRow(ctx, query, vaccinationID, petID).Scan(
		&item.ID, &item.PetID, &item.GeneratedFromID, &item.Status, &item.VaccineName, &item.CatalogMedicationID, &item.ScheduledAt, &item.AdministeredAt,
		&item.NextDueAt, &item.VetVisitID, &item.ClinicName, &item.VetName, &item.Notes, &item.RowVersion,
		&item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	attachments, err := r.listHealthAttachments(ctx, "VACCINATION", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	targets, err := r.listVaccinationTargets(ctx, item.PetID, item.ID)
	if err != nil {
		return nil, err
	}
	item.Targets = targets
	return &item, nil
}

func (r *LogRepository) GetGeneratedVaccination(ctx context.Context, petID, generatedFromID uuid.UUID) (*model.Vaccination, error) {
	const query = `SELECT id FROM vaccinations WHERE pet_id = $1 AND generated_from_id = $2 AND deleted_at IS NULL`
	var id uuid.UUID
	if err := r.db.QueryRow(ctx, query, petID, generatedFromID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return r.GetVaccination(ctx, petID, id)
}

func (r *LogRepository) ListVaccinations(ctx context.Context, in ports.ListVaccinationsQuery) (ports.ListVaccinationsResult, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeVaccinationSort(in.Sort)
	if !ok {
		return ports.ListVaccinationsResult{}, ports.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"v.pet_id = $1", "v.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("v.status = $%d", len(args)))
	}
	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf(`(
			v.vaccine_name ILIKE $%d
			OR COALESCE(v.clinic_name, '') ILIKE $%d
			OR COALESCE(v.vet_name, '') ILIKE $%d
			OR COALESCE(v.notes, '') ILIKE $%d
		)`, len(args), len(args), len(args), len(args)))
	}
	switch in.Bucket {
	case "upcoming":
		where = append(where, "v.status = 'PLANNED'")
	case "history":
		where = append(where, "v.status = 'COMPLETED'")
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("COALESCE(v.scheduled_at, v.administered_at) >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("COALESCE(v.scheduled_at, v.administered_at) <= $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(%s, v.id) %s ($%d, $%d)", orderExpr, cursorOp, len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT
			v.id, v.pet_id, v.generated_from_id, v.status, v.vaccine_name, v.catalog_medication_id, v.scheduled_at, v.administered_at, v.next_due_at,
			rv.vet_visit_id, v.clinic_name, v.vet_name, v.notes,
			COALESCE(att.cnt, 0) AS attachments_count,
			v.row_version, v.created_at, v.created_by_user_id, v.updated_at, v.updated_by_user_id,
			%s AS cursor_sort_at
		FROM vaccinations v
		LEFT JOIN LATERAL (
			SELECT rel.right_entity_id AS vet_visit_id
			FROM entity_relations rel
			JOIN vet_visits vv ON vv.id = rel.right_entity_id AND vv.pet_id = rel.pet_id AND vv.deleted_at IS NULL
			WHERE rel.pet_id = v.pet_id
			  AND rel.left_entity_type = 'VACCINATION'
			  AND rel.left_entity_id = v.id
			  AND rel.right_entity_type = 'VET_VISIT'
			LIMIT 1
		) rv ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM attachment_refs ar WHERE ar.entity_type = 'VACCINATION' AND ar.entity_id = v.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, v.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ports.ListVaccinationsResult{}, err
	}
	defer rows.Close()
	items := make([]model.VaccinationListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.VaccinationListItem
		var cursorTime time.Time
		if err := rows.Scan(
			&item.ID, &item.PetID, &item.GeneratedFromID, &item.Status, &item.VaccineName, &item.CatalogMedicationID, &item.ScheduledAt, &item.AdministeredAt,
			&item.NextDueAt, &item.VetVisitID, &item.ClinicName, &item.VetName, &item.NotesPreview, &item.AttachmentsCount,
			&item.RowVersion, &item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID, &cursorTime,
		); err != nil {
			return ports.ListVaccinationsResult{}, err
		}
		item.NotesPreview = preview(item.NotesPreview, 160)
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return ports.ListVaccinationsResult{}, err
	}
	if err := r.attachVaccinationTargets(ctx, items); err != nil {
		return ports.ListVaccinationsResult{}, err
	}
	out := ports.ListVaccinationsResult{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &ports.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateVaccination(ctx context.Context, in ports.CreateVaccinationInput) (*model.Vaccination, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO vaccinations (
			id, pet_id, generated_from_id, status, vaccine_name, catalog_medication_id, scheduled_at, administered_at, next_due_at,
			clinic_name, vet_name, notes, row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,1,NOW(),$13,NOW(),$14)
	`
	_, err = tx.Exec(ctx, query,
		in.ID, in.PetID, in.GeneratedFromID, in.Status, in.VaccineName, in.CatalogMedicationID, in.ScheduledAt, in.AdministeredAt, in.NextDueAt,
		in.ClinicName, in.VetName, in.Notes, in.CreatedBy, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VACCINATION", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := r.replaceVaccinationTargetsTx(ctx, tx, in.PetID, in.ID, in.TargetItemIDs, in.CreatedBy); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := r.replaceVisitRelationTx(ctx, tx, in.PetID, "VACCINATION", in.ID, in.VetVisitID, in.CreatedBy); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetVaccination(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateVaccination(ctx context.Context, in ports.UpdateVaccinationInput) (*model.Vaccination, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE vaccinations
		SET status = $4,
		    vaccine_name = $5,
		    catalog_medication_id = $6,
		    scheduled_at = $7,
		    administered_at = $8,
		    next_due_at = $9,
		    clinic_name = $10,
		    vet_name = $11,
		    notes = $12,
		    updated_at = NOW(),
		    updated_by_user_id = $13,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query,
		in.ID, in.PetID, in.RowVersion, in.Status, in.VaccineName, in.CatalogMedicationID, in.ScheduledAt, in.AdministeredAt,
		in.NextDueAt, in.ClinicName, in.VetName, in.Notes, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "vaccinations", in.ID, in.PetID); err != nil {
			return nil, ports.AttachmentSync{}, err
		} else if !exists {
			return nil, ports.AttachmentSync{}, ports.ErrNotFound
		}
		return nil, ports.AttachmentSync{}, ports.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VACCINATION", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if in.TargetItemIDs != nil {
		if err := r.replaceVaccinationTargetsTx(ctx, tx, in.PetID, in.ID, in.TargetItemIDs, in.UpdatedBy); err != nil {
			return nil, ports.AttachmentSync{}, err
		}
	}
	if err := r.replaceVisitRelationTx(ctx, tx, in.PetID, "VACCINATION", in.ID, in.VetVisitID, in.UpdatedBy); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetVaccination(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateGeneratedVaccinationPlan(ctx context.Context, in ports.UpdateGeneratedVaccinationPlanInput) (*model.Vaccination, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE vaccinations
		SET vaccine_name = $3,
		    catalog_medication_id = $4,
		    scheduled_at = $5,
		    updated_at = NOW(),
		    updated_by_user_id = $6,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND status = 'PLANNED' AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query, in.ID, in.PetID, in.VaccineName, in.CatalogMedicationID, in.ScheduledAt, in.UpdatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrConflict
		}
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, ports.ErrNotFound
	}
	if err := r.replaceVaccinationTargetsTx(ctx, tx, in.PetID, in.ID, in.TargetItemIDs, in.UpdatedBy); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetVaccination(ctx, in.PetID, in.ID)
}

func (r *LogRepository) DeleteVaccination(ctx context.Context, in ports.DeleteVaccinationInput) error {
	return r.softDeleteHealthEntity(ctx, "vaccinations", "VACCINATION", in.ID, in.PetID, in.RowVersion, in.DeletedBy)
}

func (r *LogRepository) GetProcedure(ctx context.Context, petID, procedureID uuid.UUID) (*model.Procedure, error) {
	const query = `
		SELECT id, pet_id, generated_from_id, status, procedure_type_item_id, title, description, catalog_medication_id, product_name, scheduled_at,
		       performed_at, next_due_at,
		       (
				SELECT rel.right_entity_id
				FROM entity_relations rel
				JOIN vet_visits vv ON vv.id = rel.right_entity_id AND vv.pet_id = rel.pet_id AND vv.deleted_at IS NULL
				WHERE rel.pet_id = p.pet_id
				  AND rel.left_entity_type = 'PROCEDURE'
				  AND rel.left_entity_id = p.id
				  AND rel.right_entity_type = 'VET_VISIT'
				LIMIT 1
		       ) AS vet_visit_id,
		       notes, row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM procedures p
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.Procedure
	var procedureTypeItemID *uuid.UUID
	err := r.db.QueryRow(ctx, query, procedureID, petID).Scan(
		&item.ID, &item.PetID, &item.GeneratedFromID, &item.Status, &procedureTypeItemID, &item.Title, &item.Description, &item.CatalogMedicationID, &item.ProductName,
		&item.ScheduledAt, &item.PerformedAt, &item.NextDueAt, &item.VetVisitID, &item.Notes, &item.RowVersion,
		&item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	attachments, err := r.listHealthAttachments(ctx, "PROCEDURE", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	if procedureTypeItemID != nil {
		typeItem, err := r.GetHealthDictionaryItem(ctx, item.PetID, *procedureTypeItemID, ports.HealthDictionaryKindProcedureType)
		if err != nil {
			return nil, err
		}
		item.ProcedureTypeItem = typeItem
	}
	return &item, nil
}

func (r *LogRepository) GetGeneratedProcedure(ctx context.Context, petID, generatedFromID uuid.UUID) (*model.Procedure, error) {
	const query = `SELECT id FROM procedures WHERE pet_id = $1 AND generated_from_id = $2 AND deleted_at IS NULL`
	var id uuid.UUID
	if err := r.db.QueryRow(ctx, query, petID, generatedFromID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return r.GetProcedure(ctx, petID, id)
}

func (r *LogRepository) ListProcedures(ctx context.Context, in ports.ListProceduresQuery) (ports.ListProceduresResult, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeProcedureSort(in.Sort)
	if !ok {
		return ports.ListProceduresResult{}, ports.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"p.pet_id = $1", "p.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("p.status = $%d", len(args)))
	}
	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf(`(
			p.title ILIKE $%d
			OR COALESCE(p.description, '') ILIKE $%d
			OR COALESCE(p.product_name, '') ILIKE $%d
			OR COALESCE(p.notes, '') ILIKE $%d
			OR EXISTS (
				SELECT 1
				FROM health_dictionary_items hdi
				WHERE hdi.id = p.procedure_type_item_id
				  AND (hdi.name ILIKE $%d OR COALESCE(hdi.code, '') ILIKE $%d)
			)
		)`, len(args), len(args), len(args), len(args), len(args), len(args)))
	}
	switch in.Bucket {
	case "planned":
		where = append(where, "p.status = 'PLANNED'")
	case "history":
		where = append(where, "p.status = 'COMPLETED'")
	}
	if in.ProcedureTypeItemID != nil {
		args = append(args, *in.ProcedureTypeItemID)
		where = append(where, fmt.Sprintf("p.procedure_type_item_id = $%d", len(args)))
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("COALESCE(p.scheduled_at, p.performed_at) >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("COALESCE(p.scheduled_at, p.performed_at) <= $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(%s, p.id) %s ($%d, $%d)", orderExpr, cursorOp, len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT
			p.id, p.pet_id, p.generated_from_id, p.status, p.procedure_type_item_id, p.title, p.description, p.catalog_medication_id, p.product_name,
			p.scheduled_at, p.performed_at, p.next_due_at, rv.vet_visit_id, p.notes,
			COALESCE(att.cnt, 0) AS attachments_count,
			p.row_version, p.created_at, p.created_by_user_id, p.updated_at, p.updated_by_user_id,
			%s AS cursor_sort_at
		FROM procedures p
		LEFT JOIN LATERAL (
			SELECT rel.right_entity_id AS vet_visit_id
			FROM entity_relations rel
			JOIN vet_visits vv ON vv.id = rel.right_entity_id AND vv.pet_id = rel.pet_id AND vv.deleted_at IS NULL
			WHERE rel.pet_id = p.pet_id
			  AND rel.left_entity_type = 'PROCEDURE'
			  AND rel.left_entity_id = p.id
			  AND rel.right_entity_type = 'VET_VISIT'
			LIMIT 1
		) rv ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM attachment_refs ar WHERE ar.entity_type = 'PROCEDURE' AND ar.entity_id = p.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, p.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ports.ListProceduresResult{}, err
	}
	defer rows.Close()
	items := make([]model.ProcedureListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.ProcedureListItem
		var cursorTime time.Time
		var procedureTypeItemID *uuid.UUID
		if err := rows.Scan(
			&item.ID, &item.PetID, &item.GeneratedFromID, &item.Status, &procedureTypeItemID, &item.Title, &item.DescriptionPreview, &item.CatalogMedicationID,
			&item.ProductName, &item.ScheduledAt, &item.PerformedAt, &item.NextDueAt, &item.VetVisitID, &item.NotesPreview,
			&item.AttachmentsCount, &item.RowVersion, &item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID, &cursorTime,
		); err != nil {
			return ports.ListProceduresResult{}, err
		}
		if procedureTypeItemID != nil {
			typeItem, err := r.GetHealthDictionaryItem(ctx, item.PetID, *procedureTypeItemID, ports.HealthDictionaryKindProcedureType)
			if err != nil {
				return ports.ListProceduresResult{}, err
			}
			item.ProcedureTypeItem = typeItem
		}
		item.DescriptionPreview = preview(item.DescriptionPreview, 160)
		item.NotesPreview = preview(item.NotesPreview, 160)
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return ports.ListProceduresResult{}, err
	}
	out := ports.ListProceduresResult{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &ports.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateProcedure(ctx context.Context, in ports.CreateProcedureInput) (*model.Procedure, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO procedures (
			id, pet_id, generated_from_id, status, procedure_type_item_id, title, description, catalog_medication_id, product_name,
			scheduled_at, performed_at, next_due_at, notes,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,1,NOW(),$14,NOW(),$15)
	`
	_, err = tx.Exec(ctx, query,
		in.ID, in.PetID, in.GeneratedFromID, in.Status, in.ProcedureTypeItemID, in.Title, in.Description, in.CatalogMedicationID, in.ProductName,
		in.ScheduledAt, in.PerformedAt, in.NextDueAt, in.Notes, in.CreatedBy, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "PROCEDURE", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := r.replaceVisitRelationTx(ctx, tx, in.PetID, "PROCEDURE", in.ID, in.VetVisitID, in.CreatedBy); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetProcedure(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateProcedure(ctx context.Context, in ports.UpdateProcedureInput) (*model.Procedure, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE procedures
		SET status = $4,
		    procedure_type_item_id = $5,
		    title = $6,
		    description = $7,
		    catalog_medication_id = $8,
		    product_name = $9,
		    scheduled_at = $10,
		    performed_at = $11,
		    next_due_at = $12,
		    notes = $13,
		    updated_at = NOW(),
		    updated_by_user_id = $14,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query,
		in.ID, in.PetID, in.RowVersion, in.Status, in.ProcedureTypeItemID, in.Title, in.Description, in.CatalogMedicationID,
		in.ProductName, in.ScheduledAt, in.PerformedAt, in.NextDueAt, in.Notes, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "procedures", in.ID, in.PetID); err != nil {
			return nil, ports.AttachmentSync{}, err
		} else if !exists {
			return nil, ports.AttachmentSync{}, ports.ErrNotFound
		}
		return nil, ports.AttachmentSync{}, ports.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "PROCEDURE", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := r.replaceVisitRelationTx(ctx, tx, in.PetID, "PROCEDURE", in.ID, in.VetVisitID, in.UpdatedBy); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetProcedure(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateGeneratedProcedurePlan(ctx context.Context, in ports.UpdateGeneratedProcedurePlanInput) (*model.Procedure, error) {
	const query = `
		UPDATE procedures
		SET procedure_type_item_id = $3,
		    title = $4,
		    description = $5,
		    catalog_medication_id = $6,
		    product_name = $7,
		    scheduled_at = $8,
		    updated_at = NOW(),
		    updated_by_user_id = $9,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND status = 'PLANNED' AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.ProcedureTypeItemID, in.Title, in.Description, in.CatalogMedicationID, in.ProductName, in.ScheduledAt, in.UpdatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrConflict
		}
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, ports.ErrNotFound
	}
	return r.GetProcedure(ctx, in.PetID, in.ID)
}

func (r *LogRepository) DeleteProcedure(ctx context.Context, in ports.DeleteProcedureInput) error {
	return r.softDeleteHealthEntity(ctx, "procedures", "PROCEDURE", in.ID, in.PetID, in.RowVersion, in.DeletedBy)
}

func (r *LogRepository) GetMedicalRecord(ctx context.Context, petID, recordID uuid.UUID) (*model.MedicalRecord, error) {
	const query = `
		SELECT id, pet_id, record_type_item_id, status, title, description, started_at, resolved_at,
		       row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM medical_records
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.MedicalRecord
	var recordTypeItemID *uuid.UUID
	err := r.db.QueryRow(ctx, query, recordID, petID).Scan(
		&item.ID, &item.PetID, &recordTypeItemID, &item.Status, &item.Title, &item.Description,
		&item.StartedAt, &item.ResolvedAt, &item.RowVersion, &item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	attachments, err := r.listHealthAttachments(ctx, "MEDICAL_RECORD", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	if recordTypeItemID != nil {
		typeItem, err := r.GetHealthDictionaryItem(ctx, item.PetID, *recordTypeItemID, ports.HealthDictionaryKindMedicalRecordType)
		if err != nil {
			return nil, err
		}
		item.RecordTypeItem = typeItem
	}
	return &item, nil
}

func (r *LogRepository) ListMedicalRecords(ctx context.Context, in ports.ListMedicalRecordsQuery) (ports.ListMedicalRecordsResult, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeMedicalRecordSort(in.Sort)
	if !ok {
		return ports.ListMedicalRecordsResult{}, ports.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"mr.pet_id = $1", "mr.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("mr.status = $%d", len(args)))
	}
	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf(`(
			mr.title ILIKE $%d
			OR COALESCE(mr.description, '') ILIKE $%d
			OR EXISTS (
				SELECT 1
				FROM health_dictionary_items hdi
				WHERE hdi.id = mr.record_type_item_id
				  AND (hdi.name ILIKE $%d OR COALESCE(hdi.code, '') ILIKE $%d)
			)
		)`, len(args), len(args), len(args), len(args)))
	}
	switch in.Bucket {
	case "active":
		where = append(where, "mr.status = 'ACTIVE'")
	case "archive":
		where = append(where, "mr.status = 'RESOLVED'")
	}
	if in.RecordTypeItemID != nil {
		args = append(args, *in.RecordTypeItemID)
		where = append(where, fmt.Sprintf("mr.record_type_item_id = $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(%s, mr.id) %s ($%d, $%d)", orderExpr, cursorOp, len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT
			mr.id, mr.pet_id, mr.record_type_item_id, mr.status, mr.title, mr.description, mr.started_at, mr.resolved_at,
			COALESCE(att.cnt, 0) AS attachments_count,
			mr.row_version, mr.created_at, mr.created_by_user_id, mr.updated_at, mr.updated_by_user_id,
			%s AS cursor_sort_at
		FROM medical_records mr
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM attachment_refs ar WHERE ar.entity_type = 'MEDICAL_RECORD' AND ar.entity_id = mr.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, mr.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ports.ListMedicalRecordsResult{}, err
	}
	defer rows.Close()
	items := make([]model.MedicalRecordListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.MedicalRecordListItem
		var cursorTime time.Time
		var recordTypeItemID *uuid.UUID
		if err := rows.Scan(
			&item.ID, &item.PetID, &recordTypeItemID, &item.Status, &item.Title, &item.DescriptionPreview, &item.StartedAt,
			&item.ResolvedAt, &item.AttachmentsCount, &item.RowVersion, &item.CreatedAt, &item.CreatedByUserID,
			&item.UpdatedAt, &item.UpdatedByUserID, &cursorTime,
		); err != nil {
			return ports.ListMedicalRecordsResult{}, err
		}
		if recordTypeItemID != nil {
			typeItem, err := r.GetHealthDictionaryItem(ctx, item.PetID, *recordTypeItemID, ports.HealthDictionaryKindMedicalRecordType)
			if err != nil {
				return ports.ListMedicalRecordsResult{}, err
			}
			item.RecordTypeItem = typeItem
		}
		item.DescriptionPreview = preview(item.DescriptionPreview, 160)
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return ports.ListMedicalRecordsResult{}, err
	}
	out := ports.ListMedicalRecordsResult{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &ports.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateMedicalRecord(ctx context.Context, in ports.CreateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO medical_records (
			id, pet_id, record_type_item_id, status, title, description, started_at, resolved_at,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,1,NOW(),$9,NOW(),$10)
	`
	_, err = tx.Exec(ctx, query, in.ID, in.PetID, in.RecordTypeItemID, in.Status, in.Title, in.Description, in.StartedAt, in.ResolvedAt, in.CreatedBy, in.UpdatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "MEDICAL_RECORD", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetMedicalRecord(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateMedicalRecord(ctx context.Context, in ports.UpdateMedicalRecordInput) (*model.MedicalRecord, ports.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE medical_records
		SET record_type_item_id = $4,
		    status = $5,
		    title = $6,
		    description = $7,
		    started_at = $8,
		    resolved_at = $9,
		    updated_at = NOW(),
		    updated_by_user_id = $10,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.RecordTypeItemID, in.Status, in.Title, in.Description, in.StartedAt, in.ResolvedAt, in.UpdatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.AttachmentSync{}, ports.ErrConflict
		}
		return nil, ports.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "medical_records", in.ID, in.PetID); err != nil {
			return nil, ports.AttachmentSync{}, err
		} else if !exists {
			return nil, ports.AttachmentSync{}, ports.ErrNotFound
		}
		return nil, ports.AttachmentSync{}, ports.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "MEDICAL_RECORD", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, ports.AttachmentSync{}, err
	}
	item, err := r.GetMedicalRecord(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) DeleteMedicalRecord(ctx context.Context, in ports.DeleteMedicalRecordInput) error {
	return r.softDeleteHealthEntity(ctx, "medical_records", "MEDICAL_RECORD", in.ID, in.PetID, in.RowVersion, in.DeletedBy)
}

func (r *LogRepository) ListCalendarDayMedicalFacts(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.CalendarDayItem, error) {
	const query = `
		SELECT item_type, entity_id, pet_id, title, subtitle, scheduled_for, status, visit_id, vaccination_id, procedure_id
		FROM (
			SELECT
				'VET_VISIT'::text AS item_type,
				vv.id AS entity_id,
				vv.pet_id,
				COALESCE(NULLIF(TRIM(vv.title), ''), 'Прием у ветеринара') AS title,
				NULLIF(TRIM(BOTH ', ' FROM CONCAT(COALESCE(vv.clinic_name, ''), CASE WHEN vv.clinic_name IS NOT NULL AND vv.vet_name IS NOT NULL THEN ', ' ELSE '' END, COALESCE(vv.vet_name, ''))), '') AS subtitle,
				COALESCE(vv.completed_at, vv.scheduled_at) AS scheduled_for,
				vv.status,
				vv.id AS visit_id,
				NULL::uuid AS vaccination_id,
				NULL::uuid AS procedure_id
			FROM vet_visits vv
			WHERE vv.pet_id = ANY($1) AND vv.deleted_at IS NULL AND vv.status = 'COMPLETED' AND COALESCE(vv.completed_at, vv.scheduled_at) >= $2 AND COALESCE(vv.completed_at, vv.scheduled_at) <= $3
			UNION ALL
			SELECT
				'VACCINATION'::text,
				v.id,
				v.pet_id,
				v.vaccine_name,
				'Вакцинация'::text,
				COALESCE(v.administered_at, v.scheduled_at),
				v.status,
				NULL::uuid,
				v.id,
				NULL::uuid
			FROM vaccinations v
			WHERE v.pet_id = ANY($1) AND v.deleted_at IS NULL AND v.status = 'COMPLETED' AND COALESCE(v.administered_at, v.scheduled_at) >= $2 AND COALESCE(v.administered_at, v.scheduled_at) <= $3
			UNION ALL
			SELECT
				'PROCEDURE'::text,
				p.id,
				p.pet_id,
				p.title,
				p.product_name,
				COALESCE(p.performed_at, p.scheduled_at),
				p.status,
				NULL::uuid,
				NULL::uuid,
				p.id
			FROM procedures p
			WHERE p.pet_id = ANY($1) AND p.deleted_at IS NULL AND p.status = 'COMPLETED' AND COALESCE(p.performed_at, p.scheduled_at) >= $2 AND COALESCE(p.performed_at, p.scheduled_at) <= $3
		) x
		ORDER BY scheduled_for ASC, pet_id ASC, item_type ASC, entity_id ASC
	`
	rows, err := r.db.Query(ctx, query, petIDs, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.CalendarDayItem, 0)
	for rows.Next() {
		var item model.CalendarDayItem
		if err := rows.Scan(&item.ItemType, &item.EntityID, &item.PetID, &item.Title, &item.Subtitle, &item.ScheduledFor, &item.Status, &item.VisitID, &item.VaccinationID, &item.ProcedureID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) listVetVisitRelatedLogs(ctx context.Context, petID, visitID uuid.UUID) ([]model.RelatedLog, error) {
	const query = `
		SELECT l.id, l.occurred_at, lt.name, l.description
		FROM entity_relations rel
		JOIN logs l ON l.id = rel.left_entity_id
		LEFT JOIN log_types lt ON lt.id = l.log_type_id
		WHERE rel.pet_id = $2
		  AND rel.left_entity_type = 'LOG'
		  AND rel.right_entity_type = 'VET_VISIT'
		  AND rel.right_entity_id = $1
		  AND l.pet_id = $2
		  AND l.deleted_at IS NULL
		ORDER BY l.occurred_at DESC, l.id DESC
	`
	rows, err := r.db.Query(ctx, query, visitID, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.RelatedLog, 0)
	for rows.Next() {
		var item model.RelatedLog
		var description *string
		if err := rows.Scan(&item.ID, &item.OccurredAt, &item.LogTypeName, &description); err != nil {
			return nil, err
		}
		item.DescriptionPreview = preview(description, 160)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) getRelatedLog(ctx context.Context, petID, logID uuid.UUID) (*model.RelatedLog, error) {
	const query = `
		SELECT l.id, l.occurred_at, lt.name, l.description
		FROM logs l
		LEFT JOIN log_types lt ON lt.id = l.log_type_id
		WHERE l.id = $1 AND l.pet_id = $2 AND l.deleted_at IS NULL
	`
	var item model.RelatedLog
	var description *string
	err := r.db.QueryRow(ctx, query, logID, petID).Scan(&item.ID, &item.OccurredAt, &item.LogTypeName, &description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	item.DescriptionPreview = preview(description, 160)
	return &item, nil
}

func (r *LogRepository) listHealthAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]model.HealthAttachment, error) {
	const query = `
		SELECT id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at
		FROM attachment_refs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY added_at ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.HealthAttachment, 0)
	for rows.Next() {
		var item model.HealthAttachment
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.FileID, &item.FileName, &item.FileType, &item.AddedByUserID, &item.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LogRepository) listVaccinationTargets(ctx context.Context, petID, vaccinationID uuid.UUID) ([]model.HealthDictionaryItem, error) {
	const query = `
		SELECT hdi.id, hdi.kind, hdi.pet_id, hdi.code, hdi.name, hdi.is_system, hdi.is_archived,
		       hdi.created_at, hdi.created_by_user_id, hdi.updated_at, hdi.updated_by_user_id
		FROM vaccination_target_links vtl
		JOIN health_dictionary_items hdi ON hdi.id = vtl.target_item_id
		WHERE vtl.pet_id = $1 AND vtl.vaccination_id = $2
		ORDER BY hdi.is_system DESC, hdi.name ASC, hdi.id ASC
	`
	rows, err := r.db.Query(ctx, query, petID, vaccinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.HealthDictionaryItem, 0)
	for rows.Next() {
		item, err := scanHealthDictionaryItem(rows)
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

func (r *LogRepository) attachVaccinationTargets(ctx context.Context, items []model.VaccinationListItem) error {
	for i := range items {
		targets, err := r.listVaccinationTargets(ctx, items[i].PetID, items[i].ID)
		if err != nil {
			return err
		}
		items[i].Targets = targets
	}
	return nil
}

func (r *LogRepository) ListPetDocuments(ctx context.Context, in ports.ListPetDocumentsQuery) (ports.ListPetDocumentsResult, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}

	args := []any{in.PetID}
	where := []string{"pet_id = $1"}
	if in.EntityType != nil {
		args = append(args, *in.EntityType)
		where = append(where, fmt.Sprintf("entity_type = $%d", len(args)))
	}
	if in.ExcludeLog {
		where = append(where, "entity_type <> 'LOG'")
	}
	if in.FileType != nil {
		args = append(args, *in.FileType)
		where = append(where, fmt.Sprintf("file_type = $%d", len(args)))
	}
	if strings.TrimSpace(in.Q) != "" {
		args = append(args, "%"+strings.TrimSpace(in.Q)+"%")
		where = append(where, fmt.Sprintf("COALESCE(file_name, '') ILIKE $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(added_at, id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT id, pet_id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at
		FROM attachment_refs
		WHERE %s
		ORDER BY added_at DESC, id DESC
		LIMIT $%d
	`, strings.Join(where, " AND "), len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ports.ListPetDocumentsResult{}, err
	}
	defer rows.Close()

	items := make([]model.PetDocument, 0, in.Limit+1)
	cursorRows := make([]struct {
		AddedAt time.Time
		ID      uuid.UUID
	}, 0, in.Limit+1)
	for rows.Next() {
		var item model.PetDocument
		if err := rows.Scan(&item.ID, &item.PetID, &item.EntityType, &item.EntityID, &item.FileID, &item.FileName, &item.FileType, &item.AddedByUserID, &item.AddedAt); err != nil {
			return ports.ListPetDocumentsResult{}, err
		}
		items = append(items, item)
		cursorRows = append(cursorRows, struct {
			AddedAt time.Time
			ID      uuid.UUID
		}{AddedAt: item.AddedAt, ID: item.ID})
	}
	if err := rows.Err(); err != nil {
		return ports.ListPetDocumentsResult{}, err
	}

	out := ports.ListPetDocumentsResult{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &ports.TimeCursor{SortAt: cursorRows[in.Limit].AddedAt, ID: cursorRows[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) GetPetDocument(ctx context.Context, petID, documentID uuid.UUID) (*model.PetDocument, error) {
	const query = `
		SELECT id, pet_id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at
		FROM attachment_refs
		WHERE id = $1 AND pet_id = $2
	`
	item := model.PetDocument{}
	err := r.db.QueryRow(ctx, query, documentID, petID).Scan(
		&item.ID,
		&item.PetID,
		&item.EntityType,
		&item.EntityID,
		&item.FileID,
		&item.FileName,
		&item.FileType,
		&item.AddedByUserID,
		&item.AddedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *LogRepository) RenamePetDocument(ctx context.Context, in ports.RenamePetDocumentInput) (*model.PetDocument, error) {
	const query = `
		UPDATE attachment_refs
		SET file_name = $3
		WHERE id = $1 AND pet_id = $2
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.FileName)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, ports.ErrNotFound
	}
	return r.GetPetDocument(ctx, in.PetID, in.ID)
}

func (r *LogRepository) replaceHealthAttachmentsTx(ctx context.Context, tx pgx.Tx, petID uuid.UUID, entityType string, entityID uuid.UUID, addedBy uuid.UUID, attachments []ports.AttachmentInput) (ports.AttachmentSync, error) {
	existingRows, err := tx.Query(ctx, `SELECT file_id FROM attachment_refs WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID)
	if err != nil {
		return ports.AttachmentSync{}, err
	}
	defer existingRows.Close()
	existing := make(map[uuid.UUID]struct{})
	for existingRows.Next() {
		var fileID uuid.UUID
		if err := existingRows.Scan(&fileID); err != nil {
			return ports.AttachmentSync{}, err
		}
		existing[fileID] = struct{}{}
	}
	if err := existingRows.Err(); err != nil {
		return ports.AttachmentSync{}, err
	}
	desired := make(map[uuid.UUID]ports.AttachmentInput, len(attachments))
	sync := ports.AttachmentSync{Add: []uuid.UUID{}, Remove: []uuid.UUID{}}
	for i := range attachments {
		att := attachments[i]
		desired[att.FileID] = att
		if _, ok := existing[att.FileID]; ok {
			_, err := tx.Exec(ctx, `UPDATE attachment_refs SET file_name = $4, file_type = $5 WHERE entity_type = $1 AND entity_id = $2 AND file_id = $3`, entityType, entityID, att.FileID, att.FileName, att.FileType)
			if err != nil {
				return ports.AttachmentSync{}, err
			}
			continue
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO attachment_refs (id, pet_id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		`, uuid.New(), petID, entityType, entityID, att.FileID, att.FileName, att.FileType, addedBy)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return ports.AttachmentSync{}, err
		}
		sync.Add = append(sync.Add, att.FileID)
	}
	for fileID := range existing {
		if _, ok := desired[fileID]; ok {
			continue
		}
		if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = $1 AND entity_id = $2 AND file_id = $3`, entityType, entityID, fileID); err != nil {
			return ports.AttachmentSync{}, err
		}
		sync.Remove = append(sync.Remove, fileID)
	}
	return sync, nil
}

func (r *LogRepository) replaceVaccinationTargetsTx(ctx context.Context, tx pgx.Tx, petID, vaccinationID uuid.UUID, targetIDs []uuid.UUID, addedBy uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM vaccination_target_links WHERE pet_id = $1 AND vaccination_id = $2`, petID, vaccinationID); err != nil {
		return err
	}
	seen := make(map[uuid.UUID]struct{}, len(targetIDs))
	for i := range targetIDs {
		targetID := targetIDs[i]
		if targetID == uuid.Nil {
			return ports.ErrConflict
		}
		if _, ok := seen[targetID]; ok {
			continue
		}
		seen[targetID] = struct{}{}
		if exists, err := r.healthDictionaryItemExistsTx(ctx, tx, petID, targetID, ports.HealthDictionaryKindVaccinationTarget); err != nil {
			return err
		} else if !exists {
			return ports.ErrNotFound
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO vaccination_target_links (vaccination_id, target_item_id, pet_id, created_by_user_id, created_at)
			VALUES ($1,$2,$3,$4,NOW())
		`, vaccinationID, targetID, petID, addedBy); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *LogRepository) replaceVisitRelationTx(ctx context.Context, tx pgx.Tx, petID uuid.UUID, entityType string, entityID uuid.UUID, visitID *uuid.UUID, addedBy uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM entity_relations
		WHERE pet_id = $1
		  AND left_entity_type = $2
		  AND left_entity_id = $3
		  AND right_entity_type = 'VET_VISIT'
	`, petID, entityType, entityID); err != nil {
		return err
	}
	if visitID == nil {
		return nil
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO entity_relations (
			id, pet_id, left_entity_type, left_entity_id, right_entity_type, right_entity_id, created_by_user_id, created_at
		) VALUES ($1,$2,$3,$4,'VET_VISIT',$5,$6,NOW())
	`, uuid.New(), petID, entityType, entityID, *visitID, addedBy); err != nil {
		if isUniqueViolation(err) {
			return ports.ErrConflict
		}
		return err
	}
	return nil
}

func (r *LogRepository) softDeleteHealthEntity(ctx context.Context, table string, entityType string, id, petID uuid.UUID, rowVersion int, deletedBy uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = NOW(),
		    deleted_by_user_id = $4,
		    updated_at = NOW(),
		    updated_by_user_id = $4,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`, table)
	cmd, err := tx.Exec(ctx, query, id, petID, rowVersion, deletedBy)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, err := r.healthEntityExistsTx(ctx, tx, table, id, petID)
		if err != nil {
			return err
		}
		if !exists {
			return ports.ErrNotFound
		}
		return ports.ErrConflict
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM entity_relations
		WHERE pet_id = $1
		  AND (
			(left_entity_type = $2 AND left_entity_id = $3)
			OR
			(right_entity_type = $2 AND right_entity_id = $3)
		  )
	`, petID, entityType, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM attachment_refs WHERE entity_type = $1 AND entity_id = $2`, entityType, id); err != nil {
		return err
	}
	if entityType == "VACCINATION" {
		if _, err := tx.Exec(ctx, `DELETE FROM vaccination_target_links WHERE pet_id = $1 AND vaccination_id = $2`, petID, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *LogRepository) healthEntityExistsTx(ctx context.Context, tx pgx.Tx, table string, id, petID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL`, table)
	var marker int
	err := tx.QueryRow(ctx, query, id, petID).Scan(&marker)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *LogRepository) healthDictionaryItemExistsTx(ctx context.Context, tx pgx.Tx, petID, itemID uuid.UUID, kind string) (bool, error) {
	const query = `
		SELECT 1
		FROM health_dictionary_items
		WHERE id = $1 AND kind = $2 AND is_archived = FALSE AND (pet_id IS NULL OR pet_id = $3)
	`
	var marker int
	err := tx.QueryRow(ctx, query, itemID, kind, petID).Scan(&marker)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

type healthDictionaryScanner interface {
	Scan(dest ...any) error
}

func scanHealthDictionaryItemRow(row healthDictionaryScanner) (model.HealthDictionaryItem, error) {
	var item model.HealthDictionaryItem
	err := row.Scan(
		&item.ID,
		&item.Kind,
		&item.PetID,
		&item.Code,
		&item.Name,
		&item.IsSystem,
		&item.IsArchived,
		&item.CreatedAt,
		&item.CreatedByUserID,
		&item.UpdatedAt,
		&item.UpdatedByUserID,
	)
	return item, err
}

func scanHealthDictionaryItem(rows pgx.Rows) (model.HealthDictionaryItem, error) {
	return scanHealthDictionaryItemRow(rows)
}

func normalizeVetVisitSort(raw string) (expr, dir, cursorOp string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "scheduled_at_desc":
		return "COALESCE(vv.scheduled_at, vv.updated_at)", "DESC", "<", true
	case "scheduled_at_asc":
		return "COALESCE(vv.scheduled_at, vv.updated_at)", "ASC", ">", true
	case "updated_at_desc":
		return "vv.updated_at", "DESC", "<", true
	default:
		return "", "", "", false
	}
}

func normalizeVaccinationSort(raw string) (expr, dir, cursorOp string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "scheduled_at_desc":
		return "COALESCE(v.scheduled_at, v.updated_at)", "DESC", "<", true
	case "scheduled_at_asc":
		return "COALESCE(v.scheduled_at, v.updated_at)", "ASC", ">", true
	case "administered_at_desc":
		return "COALESCE(v.administered_at, v.updated_at)", "DESC", "<", true
	case "updated_at_desc":
		return "v.updated_at", "DESC", "<", true
	default:
		return "", "", "", false
	}
}

func normalizeProcedureSort(raw string) (expr, dir, cursorOp string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "scheduled_at_desc":
		return "COALESCE(p.scheduled_at, p.updated_at)", "DESC", "<", true
	case "scheduled_at_asc":
		return "COALESCE(p.scheduled_at, p.updated_at)", "ASC", ">", true
	case "performed_at_desc":
		return "COALESCE(p.performed_at, p.updated_at)", "DESC", "<", true
	case "updated_at_desc":
		return "p.updated_at", "DESC", "<", true
	default:
		return "", "", "", false
	}
}

func normalizeMedicalRecordSort(raw string) (expr, dir, cursorOp string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "started_at_desc":
		return "COALESCE(mr.started_at, mr.updated_at)", "DESC", "<", true
	case "started_at_asc":
		return "COALESCE(mr.started_at, mr.updated_at)", "ASC", ">", true
	case "updated_at_desc":
		return "mr.updated_at", "DESC", "<", true
	default:
		return "", "", "", false
	}
}
