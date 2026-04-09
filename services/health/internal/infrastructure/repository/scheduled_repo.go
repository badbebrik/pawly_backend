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
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduledRepository struct {
	db *pgxpool.Pool
}

func NewScheduledRepository(db *pgxpool.Pool) *ScheduledRepository {
	return &ScheduledRepository{db: db}
}

func (r *ScheduledRepository) GetScheduledItem(ctx context.Context, petID, itemID uuid.UUID) (*model.ScheduledItem, error) {
	const query = `
		SELECT
			id,
			pet_id,
			source_type,
			source_id,
			title,
			note,
			starts_at,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id,
			deleted_at,
			deleted_by_user_id
		FROM scheduled_items
		WHERE id = $1 AND pet_id = $2 AND deleted_at IS NULL
	`
	var item model.ScheduledItem
	err := r.db.QueryRow(ctx, query, itemID, petID).Scan(
		&item.ID,
		&item.PetID,
		&item.SourceType,
		&item.SourceID,
		&item.Title,
		&item.Note,
		&item.StartsAt,
		&item.RecurrenceRule,
		&item.RecurrenceInterval,
		&item.RecurrenceUntil,
		&item.RowVersion,
		&item.CreatedAt,
		&item.CreatedByUserID,
		&item.UpdatedAt,
		&item.UpdatedByUserID,
		&item.DeletedAt,
		&item.DeletedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *ScheduledRepository) ListScheduledItems(ctx context.Context, in repo.ListScheduledItemsInput) (repo.ListScheduledItemsOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}

	args := []any{in.PetID}
	where := []string{"pet_id = $1", "deleted_at IS NULL"}
	if in.SourceType != nil {
		args = append(args, *in.SourceType)
		where = append(where, fmt.Sprintf("source_type = $%d", len(args)))
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("starts_at >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("starts_at <= $%d", len(args)))
	}
	if !in.IncludePast {
		args = append(args, time.Now().UTC())
		where = append(where, fmt.Sprintf("starts_at >= $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(starts_at, id) > ($%d, $%d)", len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)

	query := fmt.Sprintf(`
		SELECT
			id,
			pet_id,
			source_type,
			source_id,
			title,
			CASE
				WHEN note IS NULL THEN NULL
				WHEN char_length(note) <= 140 THEN note
				ELSE left(note, 140)
			END AS note_preview,
			starts_at,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id
		FROM scheduled_items
		WHERE %s
		ORDER BY starts_at ASC, id ASC
		LIMIT $%d
	`, strings.Join(where, " AND "), len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListScheduledItemsOutput{}, err
	}
	defer rows.Close()

	items := make([]model.ScheduledItemListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.ScheduledItemListItem
		if err := rows.Scan(
			&item.ID,
			&item.PetID,
			&item.SourceType,
			&item.SourceID,
			&item.Title,
			&item.NotePreview,
			&item.StartsAt,
			&item.RecurrenceRule,
			&item.RecurrenceInterval,
			&item.RecurrenceUntil,
			&item.RowVersion,
			&item.CreatedAt,
			&item.CreatedByUserID,
			&item.UpdatedAt,
			&item.UpdatedByUserID,
		); err != nil {
			return repo.ListScheduledItemsOutput{}, err
		}
		items = append(items, item)
		cursorTimes = append(cursorTimes, item.StartsAt)
	}
	if err := rows.Err(); err != nil {
		return repo.ListScheduledItemsOutput{}, err
	}

	out := repo.ListScheduledItemsOutput{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &repo.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *ScheduledRepository) CreateScheduledItem(ctx context.Context, in repo.CreateScheduledItemInput) (*model.ScheduledItem, error) {
	const query = `
		INSERT INTO scheduled_items (
			id,
			pet_id,
			source_type,
			source_id,
			title,
			note,
			starts_at,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,1,NOW(),$11,NOW(),$12
		)
	`
	_, err := r.db.Exec(ctx, query,
		in.ID,
		in.PetID,
		in.SourceType,
		in.SourceID,
		in.Title,
		in.Note,
		in.StartsAt,
		in.RecurrenceRule,
		in.RecurrenceInterval,
		in.RecurrenceUntil,
		in.CreatedBy,
		in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	return r.GetScheduledItem(ctx, in.PetID, in.ID)
}

func (r *ScheduledRepository) UpdateScheduledItem(ctx context.Context, in repo.UpdateScheduledItemInput) (*model.ScheduledItem, error) {
	const query = `
		UPDATE scheduled_items
		SET
			title = $4,
			note = $5,
			starts_at = $6,
			recurrence_rule = $7,
			recurrence_interval = $8,
			recurrence_until = $9,
			updated_at = NOW(),
			updated_by_user_id = $10,
			row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query,
		in.ID,
		in.PetID,
		in.RowVersion,
		in.Title,
		in.Note,
		in.StartsAt,
		in.RecurrenceRule,
		in.RecurrenceInterval,
		in.RecurrenceUntil,
		in.UpdatedBy,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		if _, err := r.GetScheduledItem(ctx, in.PetID, in.ID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, repo.ErrNotFound
			}
			return nil, err
		}
		return nil, repo.ErrConflict
	}
	return r.GetScheduledItem(ctx, in.PetID, in.ID)
}

func (r *ScheduledRepository) DeleteScheduledItem(ctx context.Context, in repo.DeleteScheduledItemInput) error {
	const query = `
		UPDATE scheduled_items
		SET
			deleted_at = NOW(),
			deleted_by_user_id = $4,
			updated_at = NOW(),
			updated_by_user_id = $4,
			row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.DeletedBy)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		if _, err := r.GetScheduledItem(ctx, in.PetID, in.ID); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return repo.ErrNotFound
			}
			return err
		}
		return repo.ErrConflict
	}
	return nil
}

func (r *ScheduledRepository) UpsertHealthScheduledItem(ctx context.Context, in repo.UpsertHealthScheduledItemInput) (*model.ScheduledItem, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const selectQuery = `
		SELECT id, row_version
		FROM scheduled_items
		WHERE pet_id = $1 AND source_type = $2 AND source_id = $3 AND deleted_at IS NULL
		FOR UPDATE
	`
	var itemID uuid.UUID
	var rowVersion int
	err = tx.QueryRow(ctx, selectQuery, in.PetID, in.SourceType, in.SourceID).Scan(&itemID, &rowVersion)
	switch {
	case err == nil:
		const updateQuery = `
			UPDATE scheduled_items
			SET
				title = $4,
				note = $5,
				starts_at = $6,
				updated_at = NOW(),
				updated_by_user_id = $7,
				row_version = row_version + 1
			WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
		`
		if _, err := tx.Exec(ctx, updateQuery, itemID, in.PetID, rowVersion, in.Title, in.Note, in.StartsAt, in.UpdatedByUserID); err != nil {
			return nil, err
		}
	case errors.Is(err, pgx.ErrNoRows):
		itemID = uuid.New()
		const insertQuery = `
			INSERT INTO scheduled_items (
				id,
				pet_id,
				source_type,
				source_id,
				title,
				note,
				starts_at,
				recurrence_rule,
				recurrence_interval,
				recurrence_until,
				row_version,
				created_at,
				created_by_user_id,
				updated_at,
				updated_by_user_id
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,NULL,NULL,NULL,1,NOW(),$8,NOW(),$9
			)
		`
		if _, err := tx.Exec(ctx, insertQuery, itemID, in.PetID, in.SourceType, in.SourceID, in.Title, in.Note, in.StartsAt, in.CreatedByUserID, in.UpdatedByUserID); err != nil {
			if isUniqueViolation(err) {
				return nil, repo.ErrConflict
			}
			return nil, err
		}
	default:
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetScheduledItem(ctx, in.PetID, itemID)
}

func (r *ScheduledRepository) DeleteHealthScheduledItem(ctx context.Context, in repo.DeleteHealthScheduledItemInput) error {
	const query = `
		UPDATE scheduled_items
		SET
			deleted_at = NOW(),
			deleted_by_user_id = $4,
			updated_at = NOW(),
			updated_by_user_id = $4,
			row_version = row_version + 1
		WHERE pet_id = $1 AND source_type = $2 AND source_id = $3 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, in.PetID, in.SourceType, in.SourceID, in.DeletedByUserID)
	return err
}

func (r *ScheduledRepository) ListScheduledItemOccurrences(ctx context.Context, in repo.ListScheduledItemOccurrencesInput) (repo.ListScheduledItemOccurrencesOutput, error) {
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}

	args := []any{in.PetID}
	where := []string{"o.pet_id = $1", "si.deleted_at IS NULL"}
	if in.SourceType != nil {
		args = append(args, *in.SourceType)
		where = append(where, fmt.Sprintf("si.source_type = $%d", len(args)))
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("o.scheduled_for >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("o.scheduled_for <= $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(o.scheduled_for, o.id) > ($%d, $%d)", len(args)-1, len(args)))
	}
	args = append(args, in.Limit+1)

	query := fmt.Sprintf(`
		SELECT
			o.id,
			o.scheduled_item_id,
			o.pet_id,
			o.scheduled_for,
			o.created_at,
			si.id,
			si.pet_id,
			si.source_type,
			si.source_id,
			si.title,
			si.note,
			si.starts_at,
			si.recurrence_rule,
			si.recurrence_interval,
			si.recurrence_until,
			si.row_version,
			si.created_at,
			si.created_by_user_id,
			si.updated_at,
			si.updated_by_user_id,
			si.deleted_at,
			si.deleted_by_user_id
		FROM scheduled_item_occurrences o
		JOIN scheduled_items si ON si.id = o.scheduled_item_id
		WHERE %s
		ORDER BY o.scheduled_for ASC, o.id ASC
		LIMIT $%d
	`, strings.Join(where, " AND "), len(args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return repo.ListScheduledItemOccurrencesOutput{}, err
	}
	defer rows.Close()

	items := make([]model.ScheduledItemOccurrenceListItem, 0, in.Limit+1)
	cursorTimes := make([]time.Time, 0, in.Limit+1)
	for rows.Next() {
		var item model.ScheduledItemOccurrenceListItem
		if err := rows.Scan(
			&item.ID,
			&item.ScheduledItemID,
			&item.PetID,
			&item.ScheduledFor,
			&item.CreatedAt,
			&item.Rule.ID,
			&item.Rule.PetID,
			&item.Rule.SourceType,
			&item.Rule.SourceID,
			&item.Rule.Title,
			&item.Rule.Note,
			&item.Rule.StartsAt,
			&item.Rule.RecurrenceRule,
			&item.Rule.RecurrenceInterval,
			&item.Rule.RecurrenceUntil,
			&item.Rule.RowVersion,
			&item.Rule.CreatedAt,
			&item.Rule.CreatedByUserID,
			&item.Rule.UpdatedAt,
			&item.Rule.UpdatedByUserID,
			&item.Rule.DeletedAt,
			&item.Rule.DeletedByUserID,
		); err != nil {
			return repo.ListScheduledItemOccurrencesOutput{}, err
		}
		items = append(items, item)
		cursorTimes = append(cursorTimes, item.ScheduledFor)
	}
	if err := rows.Err(); err != nil {
		return repo.ListScheduledItemOccurrencesOutput{}, err
	}

	out := repo.ListScheduledItemOccurrencesOutput{Items: items}
	if len(items) > in.Limit {
		out.NextCursor = &repo.TimeCursor{SortAt: cursorTimes[in.Limit], ID: items[in.Limit].ID}
		out.Items = items[:in.Limit]
	}
	return out, nil
}

func (r *ScheduledRepository) GetScheduledItemOccurrence(ctx context.Context, petID, occurrenceID uuid.UUID) (*model.ScheduledItemOccurrenceListItem, error) {
	const query = `
		SELECT
			o.id,
			o.scheduled_item_id,
			o.pet_id,
			o.scheduled_for,
			o.created_at,
			si.id,
			si.pet_id,
			si.source_type,
			si.source_id,
			si.title,
			si.note,
			si.starts_at,
			si.recurrence_rule,
			si.recurrence_interval,
			si.recurrence_until,
			si.row_version,
			si.created_at,
			si.created_by_user_id,
			si.updated_at,
			si.updated_by_user_id,
			si.deleted_at,
			si.deleted_by_user_id
		FROM scheduled_item_occurrences o
		JOIN scheduled_items si ON si.id = o.scheduled_item_id
		WHERE o.id = $1 AND o.pet_id = $2 AND si.deleted_at IS NULL
	`
	var item model.ScheduledItemOccurrenceListItem
	err := r.db.QueryRow(ctx, query, occurrenceID, petID).Scan(
		&item.ID,
		&item.ScheduledItemID,
		&item.PetID,
		&item.ScheduledFor,
		&item.CreatedAt,
		&item.Rule.ID,
		&item.Rule.PetID,
		&item.Rule.SourceType,
		&item.Rule.SourceID,
		&item.Rule.Title,
		&item.Rule.Note,
		&item.Rule.StartsAt,
		&item.Rule.RecurrenceRule,
		&item.Rule.RecurrenceInterval,
		&item.Rule.RecurrenceUntil,
		&item.Rule.RowVersion,
		&item.Rule.CreatedAt,
		&item.Rule.CreatedByUserID,
		&item.Rule.UpdatedAt,
		&item.Rule.UpdatedByUserID,
		&item.Rule.DeletedAt,
		&item.Rule.DeletedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *ScheduledRepository) CreateScheduledItemOccurrence(ctx context.Context, in repo.CreateScheduledItemOccurrenceInput) (*model.ScheduledItemOccurrence, error) {
	const query = `
		INSERT INTO scheduled_item_occurrences (
			id,
			scheduled_item_id,
			pet_id,
			scheduled_for,
			created_at
		) VALUES ($1,$2,$3,$4,NOW())
	`
	if _, err := r.db.Exec(ctx, query, in.ID, in.ScheduledItemID, in.PetID, in.ScheduledFor); err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	return &model.ScheduledItemOccurrence{
		ID:              in.ID,
		ScheduledItemID: in.ScheduledItemID,
		PetID:           in.PetID,
		ScheduledFor:    in.ScheduledFor,
	}, nil
}

func (r *ScheduledRepository) DeleteScheduledItemOccurrencesFrom(ctx context.Context, in repo.DeleteScheduledItemOccurrencesFromInput) error {
	const query = `
		DELETE FROM scheduled_item_occurrences
		WHERE scheduled_item_id = $1 AND scheduled_for >= $2
	`
	_, err := r.db.Exec(ctx, query, in.ScheduledItemID, in.From)
	return err
}

func (r *ScheduledRepository) CreateScheduledItemDispatch(ctx context.Context, in repo.CreateScheduledItemDispatchInput) error {
	const query = `
		INSERT INTO scheduled_item_push_dispatches (
			id,
			scheduled_item_occurrence_id,
			dispatch_key,
			created_at
		) VALUES ($1,$2,$3,NOW())
	`
	if _, err := r.db.Exec(ctx, query, in.ID, in.ScheduledItemOccurrenceID, in.DispatchKey); err != nil {
		if isUniqueViolation(err) {
			return repo.ErrConflict
		}
		return err
	}
	return nil
}

func (r *ScheduledRepository) ListDueScheduledItemOccurrences(ctx context.Context, in repo.ListDueScheduledItemOccurrencesInput) ([]model.ScheduledItemOccurrenceListItem, error) {
	if in.Limit <= 0 {
		in.Limit = 100
	}

	const query = `
		SELECT
			o.id,
			o.scheduled_item_id,
			o.pet_id,
			o.scheduled_for,
			o.created_at,
			si.id,
			si.pet_id,
			si.source_type,
			si.source_id,
			si.title,
			si.note,
			si.starts_at,
			si.recurrence_rule,
			si.recurrence_interval,
			si.recurrence_until,
			si.row_version,
			si.created_at,
			si.created_by_user_id,
			si.updated_at,
			si.updated_by_user_id,
			si.deleted_at,
			si.deleted_by_user_id
		FROM scheduled_item_occurrences o
		JOIN scheduled_items si ON si.id = o.scheduled_item_id
		LEFT JOIN scheduled_item_push_dispatches d
			ON d.scheduled_item_occurrence_id = o.id
		   AND d.dispatch_key = $2
		WHERE si.deleted_at IS NULL
		  AND o.scheduled_for <= $1
		  AND d.id IS NULL
		ORDER BY o.scheduled_for ASC, o.id ASC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, in.Before, in.DispatchKey, in.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ScheduledItemOccurrenceListItem, 0, in.Limit)
	for rows.Next() {
		var item model.ScheduledItemOccurrenceListItem
		if err := rows.Scan(
			&item.ID,
			&item.ScheduledItemID,
			&item.PetID,
			&item.ScheduledFor,
			&item.CreatedAt,
			&item.Rule.ID,
			&item.Rule.PetID,
			&item.Rule.SourceType,
			&item.Rule.SourceID,
			&item.Rule.Title,
			&item.Rule.Note,
			&item.Rule.StartsAt,
			&item.Rule.RecurrenceRule,
			&item.Rule.RecurrenceInterval,
			&item.Rule.RecurrenceUntil,
			&item.Rule.RowVersion,
			&item.Rule.CreatedAt,
			&item.Rule.CreatedByUserID,
			&item.Rule.UpdatedAt,
			&item.Rule.UpdatedByUserID,
			&item.Rule.DeletedAt,
			&item.Rule.DeletedByUserID,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ScheduledRepository) ListCalendarDayScheduledOccurrences(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error) {
	return r.listCalendarDayScheduledOccurrences(ctx, []uuid.UUID{petID}, dayStart, dayEnd)
}

func (r *ScheduledRepository) ListCalendarDayScheduledOccurrencesForPets(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error) {
	if len(petIDs) == 0 {
		return []model.ScheduledItemOccurrenceListItem{}, nil
	}
	return r.listCalendarDayScheduledOccurrences(ctx, petIDs, dayStart, dayEnd)
}

func (r *ScheduledRepository) listCalendarDayScheduledOccurrences(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItemOccurrenceListItem, error) {
	const query = `
		SELECT
			o.id,
			o.scheduled_item_id,
			o.pet_id,
			o.scheduled_for,
			o.created_at,
			si.id,
			si.pet_id,
			si.source_type,
			si.source_id,
			si.title,
			si.note,
			si.starts_at,
			si.recurrence_rule,
			si.recurrence_interval,
			si.recurrence_until,
			si.row_version,
			si.created_at,
			si.created_by_user_id,
			si.updated_at,
			si.updated_by_user_id,
			si.deleted_at,
			si.deleted_by_user_id
		FROM scheduled_item_occurrences o
		JOIN scheduled_items si ON si.id = o.scheduled_item_id
		WHERE o.pet_id = ANY($1)
		  AND si.deleted_at IS NULL
		  AND o.scheduled_for >= $2
		  AND o.scheduled_for <= $3
		ORDER BY o.scheduled_for ASC, o.id ASC
	`
	rows, err := r.db.Query(ctx, query, petIDs, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ScheduledItemOccurrenceListItem, 0)
	for rows.Next() {
		var item model.ScheduledItemOccurrenceListItem
		if err := rows.Scan(
			&item.ID,
			&item.ScheduledItemID,
			&item.PetID,
			&item.ScheduledFor,
			&item.CreatedAt,
			&item.Rule.ID,
			&item.Rule.PetID,
			&item.Rule.SourceType,
			&item.Rule.SourceID,
			&item.Rule.Title,
			&item.Rule.Note,
			&item.Rule.StartsAt,
			&item.Rule.RecurrenceRule,
			&item.Rule.RecurrenceInterval,
			&item.Rule.RecurrenceUntil,
			&item.Rule.RowVersion,
			&item.Rule.CreatedAt,
			&item.Rule.CreatedByUserID,
			&item.Rule.UpdatedAt,
			&item.Rule.UpdatedByUserID,
			&item.Rule.DeletedAt,
			&item.Rule.DeletedByUserID,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
