package repository

import (
	"context"
	"errors"
	"fmt"
	"health/internal/model"
	repo "health/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *LogRepository) GetVetVisit(ctx context.Context, petID, visitID uuid.UUID, includeRelatedLogs bool) (*model.VetVisit, error) {
	const query = `
		SELECT
			id,
			pet_id,
			status,
			visit_type,
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
			return nil, repo.ErrNotFound
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

func (r *LogRepository) ListVetVisits(ctx context.Context, in repo.ListVetVisitsInput) (repo.ListVetVisitsOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeVetVisitSort(in.Sort)
	if !ok {
		return repo.ListVetVisitsOutput{}, repo.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"vv.pet_id = $1", "vv.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("vv.status = $%d", len(args)))
	}
	switch in.Bucket {
	case "upcoming":
		where = append(where, "vv.status = 'PLANNED'")
	case "history":
		where = append(where, "vv.status IN ('COMPLETED', 'CANCELLED')")
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
			SELECT COUNT(1)::int AS cnt FROM vet_visit_log_refs ref WHERE ref.vet_visit_id = vv.id
		) rl ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM health_attachment_refs har WHERE har.entity_type = 'VET_VISIT' AND har.entity_id = vv.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, vv.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListVetVisitsOutput{}, err
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
			return repo.ListVetVisitsOutput{}, err
		}
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return repo.ListVetVisitsOutput{}, err
	}

	out := repo.ListVetVisitsOutput{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &repo.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateVetVisit(ctx context.Context, in repo.CreateVetVisitInput) (*model.VetVisit, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO vet_visits (
			id, pet_id, status, visit_type, scheduled_at, completed_at, reason_text, result_text, clinic_name, vet_name,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,1,NOW(),$11,NOW(),$12)
	`
	_, err = tx.Exec(ctx, query,
		in.ID, in.PetID, in.Status, in.VisitType, in.ScheduledAt, in.CompletedAt, in.ReasonText, in.ResultText, in.ClinicName, in.VetName,
		in.CreatedBy, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VET_VISIT", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetVetVisit(ctx, in.PetID, in.ID, true)
	return item, sync, err
}

func (r *LogRepository) UpdateVetVisit(ctx context.Context, in repo.UpdateVetVisitInput) (*model.VetVisit, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE vet_visits
		SET status = $4,
		    visit_type = $5,
		    scheduled_at = $6,
		    completed_at = $7,
		    reason_text = $8,
		    result_text = $9,
		    clinic_name = $10,
		    vet_name = $11,
		    updated_at = NOW(),
		    updated_by_user_id = $12,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query,
		in.ID, in.PetID, in.RowVersion, in.Status, in.VisitType, in.ScheduledAt, in.CompletedAt, in.ReasonText, in.ResultText, in.ClinicName, in.VetName, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "vet_visits", in.ID, in.PetID); err != nil {
			return nil, repo.AttachmentSync{}, err
		} else if !exists {
			return nil, repo.AttachmentSync{}, repo.ErrNotFound
		}
		return nil, repo.AttachmentSync{}, repo.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VET_VISIT", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetVetVisit(ctx, in.PetID, in.ID, true)
	return item, sync, err
}

func (r *LogRepository) DeleteVetVisit(ctx context.Context, in repo.DeleteVetVisitInput) error {
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
			return repo.ErrNotFound
		}
		return repo.ErrConflict
	}
	if _, err := tx.Exec(ctx, `UPDATE vaccinations SET vet_visit_id = NULL, updated_at = NOW(), updated_by_user_id = $2, row_version = row_version + 1 WHERE pet_id = $1 AND vet_visit_id = $3 AND deleted_at IS NULL`, in.PetID, in.DeletedBy, in.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE procedures SET vet_visit_id = NULL, updated_at = NOW(), updated_by_user_id = $2, row_version = row_version + 1 WHERE pet_id = $1 AND vet_visit_id = $3 AND deleted_at IS NULL`, in.PetID, in.DeletedBy, in.ID); err != nil {
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
		return nil, repo.ErrNotFound
	}
	if exists, _, err := r.lookupLogStateTx(ctx, tx, petID, logID); err != nil {
		return nil, err
	} else if !exists {
		return nil, repo.ErrNotFound
	}
	const query = `INSERT INTO vet_visit_log_refs (id, vet_visit_id, log_id, added_by_user_id, added_at) VALUES ($1,$2,$3,$4,NOW())`
	if _, err := tx.Exec(ctx, query, uuid.New(), visitID, logID, addedBy); err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
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
		DELETE FROM vet_visit_log_refs ref
		USING vet_visits vv
		WHERE ref.vet_visit_id = vv.id
		  AND vv.id = $1
		  AND vv.pet_id = $2
		  AND vv.deleted_at IS NULL
		  AND ref.log_id = $3
	`
	cmd, err := r.db.Exec(ctx, query, visitID, petID, logID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *LogRepository) GetVaccination(ctx context.Context, petID, vaccinationID uuid.UUID) (*model.Vaccination, error) {
	const query = `
		SELECT id, pet_id, status, vaccine_name, catalog_medication_id, scheduled_at, administered_at, next_due_at, vet_visit_id,
		       clinic_name, vet_name, notes, row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM vaccinations
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.Vaccination
	err := r.db.QueryRow(ctx, query, vaccinationID, petID).Scan(
		&item.ID, &item.PetID, &item.Status, &item.VaccineName, &item.CatalogMedicationID, &item.ScheduledAt, &item.AdministeredAt,
		&item.NextDueAt, &item.VetVisitID, &item.ClinicName, &item.VetName, &item.Notes, &item.RowVersion,
		&item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	attachments, err := r.listHealthAttachments(ctx, "VACCINATION", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	return &item, nil
}

func (r *LogRepository) ListVaccinations(ctx context.Context, in repo.ListVaccinationsInput) (repo.ListVaccinationsOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeVaccinationSort(in.Sort)
	if !ok {
		return repo.ListVaccinationsOutput{}, repo.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"v.pet_id = $1", "v.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("v.status = $%d", len(args)))
	}
	switch in.Bucket {
	case "planned":
		where = append(where, "v.status = 'PLANNED'")
	case "history":
		where = append(where, "v.status IN ('DONE', 'CANCELLED')")
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
			v.id, v.pet_id, v.status, v.vaccine_name, v.catalog_medication_id, v.scheduled_at, v.administered_at, v.next_due_at,
			v.vet_visit_id, v.clinic_name, v.vet_name, v.notes,
			COALESCE(att.cnt, 0) AS attachments_count,
			v.row_version, v.created_at, v.created_by_user_id, v.updated_at, v.updated_by_user_id,
			%s AS cursor_sort_at
		FROM vaccinations v
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM health_attachment_refs har WHERE har.entity_type = 'VACCINATION' AND har.entity_id = v.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, v.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListVaccinationsOutput{}, err
	}
	defer rows.Close()
	items := make([]model.VaccinationListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.VaccinationListItem
		var cursorTime time.Time
		if err := rows.Scan(
			&item.ID, &item.PetID, &item.Status, &item.VaccineName, &item.CatalogMedicationID, &item.ScheduledAt, &item.AdministeredAt,
			&item.NextDueAt, &item.VetVisitID, &item.ClinicName, &item.VetName, &item.NotesPreview, &item.AttachmentsCount,
			&item.RowVersion, &item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID, &cursorTime,
		); err != nil {
			return repo.ListVaccinationsOutput{}, err
		}
		item.NotesPreview = preview(item.NotesPreview, 160)
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return repo.ListVaccinationsOutput{}, err
	}
	out := repo.ListVaccinationsOutput{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &repo.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateVaccination(ctx context.Context, in repo.CreateVaccinationInput) (*model.Vaccination, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO vaccinations (
			id, pet_id, status, vaccine_name, catalog_medication_id, scheduled_at, administered_at, next_due_at, vet_visit_id,
			clinic_name, vet_name, notes, source_vaccination_id,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,1,NOW(),$14,NOW(),$15)
	`
	_, err = tx.Exec(ctx, query,
		in.ID, in.PetID, in.Status, in.VaccineName, in.CatalogMedicationID, in.ScheduledAt, in.AdministeredAt, in.NextDueAt, in.VetVisitID,
		in.ClinicName, in.VetName, in.Notes, in.SourceVaccinationID, in.CreatedBy, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VACCINATION", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetVaccination(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateVaccination(ctx context.Context, in repo.UpdateVaccinationInput) (*model.Vaccination, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
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
		    vet_visit_id = $10,
		    clinic_name = $11,
		    vet_name = $12,
		    notes = $13,
		    updated_at = NOW(),
		    updated_by_user_id = $14,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query,
		in.ID, in.PetID, in.RowVersion, in.Status, in.VaccineName, in.CatalogMedicationID, in.ScheduledAt, in.AdministeredAt,
		in.NextDueAt, in.VetVisitID, in.ClinicName, in.VetName, in.Notes, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "vaccinations", in.ID, in.PetID); err != nil {
			return nil, repo.AttachmentSync{}, err
		} else if !exists {
			return nil, repo.AttachmentSync{}, repo.ErrNotFound
		}
		return nil, repo.AttachmentSync{}, repo.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "VACCINATION", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetVaccination(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) DeleteVaccination(ctx context.Context, in repo.DeleteVaccinationInput) error {
	return r.softDeleteHealthEntity(ctx, "vaccinations", in.ID, in.PetID, in.RowVersion, in.DeletedBy)
}

func (r *LogRepository) HasPlannedVaccinationFromSource(ctx context.Context, petID, sourceVaccinationID uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM vaccinations WHERE pet_id = $1 AND source_vaccination_id = $2 AND status = 'PLANNED' AND deleted_at IS NULL`
	var marker int
	err := r.db.QueryRow(ctx, query, petID, sourceVaccinationID).Scan(&marker)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *LogRepository) GetProcedure(ctx context.Context, petID, procedureID uuid.UUID) (*model.Procedure, error) {
	const query = `
		SELECT id, pet_id, status, procedure_type, title, description, catalog_medication_id, product_name, scheduled_at,
		       performed_at, next_due_at, vet_visit_id, notes, row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM procedures
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.Procedure
	err := r.db.QueryRow(ctx, query, procedureID, petID).Scan(
		&item.ID, &item.PetID, &item.Status, &item.ProcedureType, &item.Title, &item.Description, &item.CatalogMedicationID, &item.ProductName,
		&item.ScheduledAt, &item.PerformedAt, &item.NextDueAt, &item.VetVisitID, &item.Notes, &item.RowVersion,
		&item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	attachments, err := r.listHealthAttachments(ctx, "PROCEDURE", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	return &item, nil
}

func (r *LogRepository) ListProcedures(ctx context.Context, in repo.ListProceduresInput) (repo.ListProceduresOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeProcedureSort(in.Sort)
	if !ok {
		return repo.ListProceduresOutput{}, repo.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"p.pet_id = $1", "p.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("p.status = $%d", len(args)))
	}
	switch in.Bucket {
	case "planned":
		where = append(where, "p.status = 'PLANNED'")
	case "history":
		where = append(where, "p.status IN ('DONE', 'CANCELLED')")
	}
	if in.ProcedureType != nil {
		args = append(args, *in.ProcedureType)
		where = append(where, fmt.Sprintf("p.procedure_type = $%d", len(args)))
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
			p.id, p.pet_id, p.status, p.procedure_type, p.title, p.description, p.catalog_medication_id, p.product_name,
			p.scheduled_at, p.performed_at, p.next_due_at, p.vet_visit_id, p.notes,
			COALESCE(att.cnt, 0) AS attachments_count,
			p.row_version, p.created_at, p.created_by_user_id, p.updated_at, p.updated_by_user_id,
			%s AS cursor_sort_at
		FROM procedures p
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM health_attachment_refs har WHERE har.entity_type = 'PROCEDURE' AND har.entity_id = p.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, p.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListProceduresOutput{}, err
	}
	defer rows.Close()
	items := make([]model.ProcedureListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.ProcedureListItem
		var cursorTime time.Time
		if err := rows.Scan(
			&item.ID, &item.PetID, &item.Status, &item.ProcedureType, &item.Title, &item.DescriptionPreview, &item.CatalogMedicationID,
			&item.ProductName, &item.ScheduledAt, &item.PerformedAt, &item.NextDueAt, &item.VetVisitID, &item.NotesPreview,
			&item.AttachmentsCount, &item.RowVersion, &item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID, &cursorTime,
		); err != nil {
			return repo.ListProceduresOutput{}, err
		}
		item.DescriptionPreview = preview(item.DescriptionPreview, 160)
		item.NotesPreview = preview(item.NotesPreview, 160)
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return repo.ListProceduresOutput{}, err
	}
	out := repo.ListProceduresOutput{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &repo.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateProcedure(ctx context.Context, in repo.CreateProcedureInput) (*model.Procedure, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO procedures (
			id, pet_id, status, procedure_type, title, description, catalog_medication_id, product_name,
			scheduled_at, performed_at, next_due_at, vet_visit_id, notes, source_procedure_id,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,1,NOW(),$15,NOW(),$16)
	`
	_, err = tx.Exec(ctx, query,
		in.ID, in.PetID, in.Status, in.ProcedureType, in.Title, in.Description, in.CatalogMedicationID, in.ProductName,
		in.ScheduledAt, in.PerformedAt, in.NextDueAt, in.VetVisitID, in.Notes, in.SourceProcedureID, in.CreatedBy, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "PROCEDURE", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetProcedure(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateProcedure(ctx context.Context, in repo.UpdateProcedureInput) (*model.Procedure, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE procedures
		SET status = $4,
		    procedure_type = $5,
		    title = $6,
		    description = $7,
		    catalog_medication_id = $8,
		    product_name = $9,
		    scheduled_at = $10,
		    performed_at = $11,
		    next_due_at = $12,
		    vet_visit_id = $13,
		    notes = $14,
		    updated_at = NOW(),
		    updated_by_user_id = $15,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query,
		in.ID, in.PetID, in.RowVersion, in.Status, in.ProcedureType, in.Title, in.Description, in.CatalogMedicationID,
		in.ProductName, in.ScheduledAt, in.PerformedAt, in.NextDueAt, in.VetVisitID, in.Notes, in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "procedures", in.ID, in.PetID); err != nil {
			return nil, repo.AttachmentSync{}, err
		} else if !exists {
			return nil, repo.AttachmentSync{}, repo.ErrNotFound
		}
		return nil, repo.AttachmentSync{}, repo.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "PROCEDURE", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetProcedure(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) DeleteProcedure(ctx context.Context, in repo.DeleteProcedureInput) error {
	return r.softDeleteHealthEntity(ctx, "procedures", in.ID, in.PetID, in.RowVersion, in.DeletedBy)
}

func (r *LogRepository) HasPlannedProcedureFromSource(ctx context.Context, petID, sourceProcedureID uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM procedures WHERE pet_id = $1 AND source_procedure_id = $2 AND status = 'PLANNED' AND deleted_at IS NULL`
	var marker int
	err := r.db.QueryRow(ctx, query, petID, sourceProcedureID).Scan(&marker)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *LogRepository) GetMedicalRecord(ctx context.Context, petID, recordID uuid.UUID) (*model.MedicalRecord, error) {
	const query = `
		SELECT id, pet_id, record_type, status, title, description, started_at, resolved_at,
		       row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		FROM medical_records
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.MedicalRecord
	err := r.db.QueryRow(ctx, query, recordID, petID).Scan(
		&item.ID, &item.PetID, &item.RecordType, &item.Status, &item.Title, &item.Description,
		&item.StartedAt, &item.ResolvedAt, &item.RowVersion, &item.CreatedAt, &item.CreatedByUserID, &item.UpdatedAt, &item.UpdatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	attachments, err := r.listHealthAttachments(ctx, "MEDICAL_RECORD", item.ID)
	if err != nil {
		return nil, err
	}
	item.Attachments = attachments
	return &item, nil
}

func (r *LogRepository) ListMedicalRecords(ctx context.Context, in repo.ListMedicalRecordsInput) (repo.ListMedicalRecordsOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	orderExpr, orderDir, cursorOp, ok := normalizeMedicalRecordSort(in.Sort)
	if !ok {
		return repo.ListMedicalRecordsOutput{}, repo.ErrConflict
	}
	args := []any{in.PetID}
	where := []string{"mr.pet_id = $1", "mr.deleted_at IS NULL"}
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("mr.status = $%d", len(args)))
	}
	switch in.Bucket {
	case "active":
		where = append(where, "mr.status = 'ACTIVE'")
	case "archive":
		where = append(where, "mr.status = 'RESOLVED'")
	}
	if in.RecordType != nil {
		args = append(args, *in.RecordType)
		where = append(where, fmt.Sprintf("mr.record_type = $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(%s, mr.id) %s ($%d, $%d)", orderExpr, cursorOp, len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)
	query := fmt.Sprintf(`
		SELECT
			mr.id, mr.pet_id, mr.record_type, mr.status, mr.title, mr.description, mr.started_at, mr.resolved_at,
			COALESCE(att.cnt, 0) AS attachments_count,
			mr.row_version, mr.created_at, mr.created_by_user_id, mr.updated_at, mr.updated_by_user_id,
			%s AS cursor_sort_at
		FROM medical_records mr
		LEFT JOIN LATERAL (
			SELECT COUNT(1)::int AS cnt FROM health_attachment_refs har WHERE har.entity_type = 'MEDICAL_RECORD' AND har.entity_id = mr.id
		) att ON TRUE
		WHERE %s
		ORDER BY %s %s, mr.id %s
		LIMIT $%d
	`, orderExpr, strings.Join(where, " AND "), orderExpr, orderDir, orderDir, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListMedicalRecordsOutput{}, err
	}
	defer rows.Close()
	items := make([]model.MedicalRecordListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.MedicalRecordListItem
		var cursorTime time.Time
		if err := rows.Scan(
			&item.ID, &item.PetID, &item.RecordType, &item.Status, &item.Title, &item.DescriptionPreview, &item.StartedAt,
			&item.ResolvedAt, &item.AttachmentsCount, &item.RowVersion, &item.CreatedAt, &item.CreatedByUserID,
			&item.UpdatedAt, &item.UpdatedByUserID, &cursorTime,
		); err != nil {
			return repo.ListMedicalRecordsOutput{}, err
		}
		item.DescriptionPreview = preview(item.DescriptionPreview, 160)
		items = append(items, item)
		cursorTimes = append(cursorTimes, cursorTime)
	}
	if err := rows.Err(); err != nil {
		return repo.ListMedicalRecordsOutput{}, err
	}
	out := repo.ListMedicalRecordsOutput{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &repo.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *LogRepository) CreateMedicalRecord(ctx context.Context, in repo.CreateMedicalRecordInput) (*model.MedicalRecord, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		INSERT INTO medical_records (
			id, pet_id, record_type, status, title, description, started_at, resolved_at,
			row_version, created_at, created_by_user_id, updated_at, updated_by_user_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,1,NOW(),$9,NOW(),$10)
	`
	_, err = tx.Exec(ctx, query, in.ID, in.PetID, in.RecordType, in.Status, in.Title, in.Description, in.StartedAt, in.ResolvedAt, in.CreatedBy, in.UpdatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "MEDICAL_RECORD", in.ID, in.CreatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetMedicalRecord(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) UpdateMedicalRecord(ctx context.Context, in repo.UpdateMedicalRecordInput) (*model.MedicalRecord, repo.AttachmentSync, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	const query = `
		UPDATE medical_records
		SET record_type = $4,
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
	cmd, err := tx.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.RecordType, in.Status, in.Title, in.Description, in.StartedAt, in.ResolvedAt, in.UpdatedBy)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.AttachmentSync{}, repo.ErrConflict
		}
		return nil, repo.AttachmentSync{}, err
	}
	if cmd.RowsAffected() == 0 {
		if exists, err := r.healthEntityExistsTx(ctx, tx, "medical_records", in.ID, in.PetID); err != nil {
			return nil, repo.AttachmentSync{}, err
		} else if !exists {
			return nil, repo.AttachmentSync{}, repo.ErrNotFound
		}
		return nil, repo.AttachmentSync{}, repo.ErrConflict
	}
	sync, err := r.replaceHealthAttachmentsTx(ctx, tx, in.PetID, "MEDICAL_RECORD", in.ID, in.UpdatedBy, in.Attachments)
	if err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, repo.AttachmentSync{}, err
	}
	item, err := r.GetMedicalRecord(ctx, in.PetID, in.ID)
	return item, sync, err
}

func (r *LogRepository) DeleteMedicalRecord(ctx context.Context, in repo.DeleteMedicalRecordInput) error {
	return r.softDeleteHealthEntity(ctx, "medical_records", in.ID, in.PetID, in.RowVersion, in.DeletedBy)
}

func (r *LogRepository) ListCalendarDayItems(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.CalendarDayItem, error) {
	const query = `
		SELECT item_type, entity_id, title, subtitle, scheduled_for, status, visit_id, vaccination_id, procedure_id
		FROM (
			SELECT
				'VET_VISIT'::text AS item_type,
				vv.id AS entity_id,
				'Прием у ветеринара'::text AS title,
				NULLIF(TRIM(BOTH ', ' FROM CONCAT(COALESCE(vv.clinic_name, ''), CASE WHEN vv.clinic_name IS NOT NULL AND vv.vet_name IS NOT NULL THEN ', ' ELSE '' END, COALESCE(vv.vet_name, ''))), '') AS subtitle,
				vv.scheduled_at AS scheduled_for,
				vv.status,
				vv.id AS visit_id,
				NULL::uuid AS vaccination_id,
				NULL::uuid AS procedure_id
			FROM vet_visits vv
			WHERE vv.pet_id = $1 AND vv.deleted_at IS NULL AND vv.status = 'PLANNED' AND vv.scheduled_at >= $2 AND vv.scheduled_at <= $3
			UNION ALL
			SELECT
				'VACCINATION'::text,
				v.id,
				v.vaccine_name,
				'Вакцинация'::text,
				v.scheduled_at,
				v.status,
				NULL::uuid,
				v.id,
				NULL::uuid
			FROM vaccinations v
			WHERE v.pet_id = $1 AND v.deleted_at IS NULL AND v.status = 'PLANNED' AND v.scheduled_at >= $2 AND v.scheduled_at <= $3
			UNION ALL
			SELECT
				'PROCEDURE'::text,
				p.id,
				p.title,
				p.product_name,
				p.scheduled_at,
				p.status,
				NULL::uuid,
				NULL::uuid,
				p.id
			FROM procedures p
			WHERE p.pet_id = $1 AND p.deleted_at IS NULL AND p.status = 'PLANNED' AND p.scheduled_at >= $2 AND p.scheduled_at <= $3
		) x
		ORDER BY scheduled_for ASC, item_type ASC, entity_id ASC
	`
	rows, err := r.db.Query(ctx, query, petID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.CalendarDayItem, 0)
	for rows.Next() {
		var item model.CalendarDayItem
		if err := rows.Scan(&item.ItemType, &item.EntityID, &item.Title, &item.Subtitle, &item.ScheduledFor, &item.Status, &item.VisitID, &item.VaccinationID, &item.ProcedureID); err != nil {
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
		SELECT l.id, l.occurred_at, lt.name, l.description, l.source
		FROM vet_visit_log_refs ref
		JOIN logs l ON l.id = ref.log_id
		LEFT JOIN log_types lt ON lt.id = l.log_type_id
		WHERE ref.vet_visit_id = $1 AND l.pet_id = $2 AND l.deleted_at IS NULL
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
		if err := rows.Scan(&item.ID, &item.OccurredAt, &item.LogTypeName, &description, &item.Source); err != nil {
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
		SELECT l.id, l.occurred_at, lt.name, l.description, l.source
		FROM logs l
		LEFT JOIN log_types lt ON lt.id = l.log_type_id
		WHERE l.id = $1 AND l.pet_id = $2 AND l.deleted_at IS NULL
	`
	var item model.RelatedLog
	var description *string
	err := r.db.QueryRow(ctx, query, logID, petID).Scan(&item.ID, &item.OccurredAt, &item.LogTypeName, &description, &item.Source)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	item.DescriptionPreview = preview(description, 160)
	return &item, nil
}

func (r *LogRepository) listHealthAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]model.HealthAttachment, error) {
	const query = `
		SELECT id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at
		FROM health_attachment_refs
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

func (r *LogRepository) replaceHealthAttachmentsTx(ctx context.Context, tx pgx.Tx, petID uuid.UUID, entityType string, entityID uuid.UUID, addedBy uuid.UUID, attachments []repo.AttachmentInput) (repo.AttachmentSync, error) {
	existingRows, err := tx.Query(ctx, `SELECT file_id FROM health_attachment_refs WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID)
	if err != nil {
		return repo.AttachmentSync{}, err
	}
	defer existingRows.Close()
	existing := make(map[uuid.UUID]struct{})
	for existingRows.Next() {
		var fileID uuid.UUID
		if err := existingRows.Scan(&fileID); err != nil {
			return repo.AttachmentSync{}, err
		}
		existing[fileID] = struct{}{}
	}
	if err := existingRows.Err(); err != nil {
		return repo.AttachmentSync{}, err
	}
	desired := make(map[uuid.UUID]repo.AttachmentInput, len(attachments))
	sync := repo.AttachmentSync{Add: []uuid.UUID{}, Remove: []uuid.UUID{}}
	for i := range attachments {
		att := attachments[i]
		desired[att.FileID] = att
		if _, ok := existing[att.FileID]; ok {
			_, err := tx.Exec(ctx, `UPDATE health_attachment_refs SET file_name = $4, file_type = $5 WHERE entity_type = $1 AND entity_id = $2 AND file_id = $3`, entityType, entityID, att.FileID, att.FileName, att.FileType)
			if err != nil {
				return repo.AttachmentSync{}, err
			}
			continue
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO health_attachment_refs (id, pet_id, entity_type, entity_id, file_id, file_name, file_type, added_by_user_id, added_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		`, uuid.New(), petID, entityType, entityID, att.FileID, att.FileName, att.FileType, addedBy)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return repo.AttachmentSync{}, err
		}
		sync.Add = append(sync.Add, att.FileID)
	}
	for fileID := range existing {
		if _, ok := desired[fileID]; ok {
			continue
		}
		if _, err := tx.Exec(ctx, `DELETE FROM health_attachment_refs WHERE entity_type = $1 AND entity_id = $2 AND file_id = $3`, entityType, entityID, fileID); err != nil {
			return repo.AttachmentSync{}, err
		}
		sync.Remove = append(sync.Remove, fileID)
	}
	return sync, nil
}

func (r *LogRepository) softDeleteHealthEntity(ctx context.Context, table string, id, petID uuid.UUID, rowVersion int, deletedBy uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = NOW(),
		    deleted_by_user_id = $4,
		    updated_at = NOW(),
		    updated_by_user_id = $4,
		    row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`, table)
	cmd, err := r.db.Exec(ctx, query, id, petID, rowVersion, deletedBy)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, err := r.healthEntityExists(ctx, table, id, petID)
		if err != nil {
			return err
		}
		if !exists {
			return repo.ErrNotFound
		}
		return repo.ErrConflict
	}
	return nil
}

func (r *LogRepository) healthEntityExists(ctx context.Context, table string, id, petID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL`, table)
	var marker int
	err := r.db.QueryRow(ctx, query, id, petID).Scan(&marker)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
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

var _ repo.HealthRepository = (*LogRepository)(nil)
