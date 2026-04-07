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
			scheduled_for,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			status,
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
		&item.ScheduledFor,
		&item.RecurrenceRule,
		&item.RecurrenceInterval,
		&item.RecurrenceUntil,
		&item.Status,
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
	if in.Status != nil {
		args = append(args, *in.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if in.SourceType != nil {
		args = append(args, *in.SourceType)
		where = append(where, fmt.Sprintf("source_type = $%d", len(args)))
	}
	if in.DateFrom != nil {
		args = append(args, *in.DateFrom)
		where = append(where, fmt.Sprintf("scheduled_for >= $%d", len(args)))
	}
	if in.DateTo != nil {
		args = append(args, *in.DateTo)
		where = append(where, fmt.Sprintf("scheduled_for <= $%d", len(args)))
	}
	if !in.IncludePast {
		args = append(args, time.Now().UTC())
		where = append(where, fmt.Sprintf("scheduled_for >= $%d", len(args)))
	}
	if in.Cursor != nil {
		args = append(args, in.Cursor.SortAt, in.Cursor.ID)
		where = append(where, fmt.Sprintf("(scheduled_for, id) > ($%d, $%d)", len(args)-1, len(args)))
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
			scheduled_for,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			status,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id
		FROM scheduled_items
		WHERE %s
		ORDER BY scheduled_for ASC, id ASC
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
			&item.ScheduledFor,
			&item.RecurrenceRule,
			&item.RecurrenceInterval,
			&item.RecurrenceUntil,
			&item.Status,
			&item.RowVersion,
			&item.CreatedAt,
			&item.CreatedByUserID,
			&item.UpdatedAt,
			&item.UpdatedByUserID,
		); err != nil {
			return repo.ListScheduledItemsOutput{}, err
		}
		items = append(items, item)
		cursorTimes = append(cursorTimes, item.ScheduledFor)
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
			scheduled_for,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			status,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1,NOW(),$12,NOW(),$13
		)
	`
	_, err := r.db.Exec(ctx, query,
		in.ID,
		in.PetID,
		in.SourceType,
		in.SourceID,
		in.Title,
		in.Note,
		in.ScheduledFor,
		in.RecurrenceRule,
		in.RecurrenceInterval,
		in.RecurrenceUntil,
		in.Status,
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
			scheduled_for = $6,
			recurrence_rule = $7,
			recurrence_interval = $8,
			recurrence_until = $9,
			status = $10,
			updated_at = NOW(),
			updated_by_user_id = $11,
			row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query,
		in.ID,
		in.PetID,
		in.RowVersion,
		in.Title,
		in.Note,
		in.ScheduledFor,
		in.RecurrenceRule,
		in.RecurrenceInterval,
		in.RecurrenceUntil,
		in.Status,
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

func (r *ScheduledRepository) CompleteScheduledItem(ctx context.Context, in repo.CompleteScheduledItemInput) (*model.ScheduledItem, error) {
	const query = `
		UPDATE scheduled_items
		SET
			status = 'DONE',
			updated_at = NOW(),
			updated_by_user_id = $4,
			row_version = row_version + 1
		WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, in.ID, in.PetID, in.RowVersion, in.UpdatedBy)
	if err != nil {
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
		WHERE pet_id = $1
		  AND source_type = $2
		  AND source_id = $3
		  AND deleted_at IS NULL
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
				scheduled_for = $6,
				status = $7,
				updated_at = NOW(),
				updated_by_user_id = $8,
				row_version = row_version + 1
			WHERE id = $1 AND pet_id = $2 AND row_version = $3 AND deleted_at IS NULL
		`
		_, err = tx.Exec(ctx, updateQuery, itemID, in.PetID, rowVersion, in.Title, in.Note, in.ScheduledFor, in.Status, in.UpdatedByUserID)
		if err != nil {
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
				scheduled_for,
				recurrence_rule,
				recurrence_interval,
				recurrence_until,
				status,
				row_version,
				created_at,
				created_by_user_id,
				updated_at,
				updated_by_user_id
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,NULL,NULL,NULL,$8,1,NOW(),$9,NOW(),$10
			)
		`
		_, err = tx.Exec(ctx, insertQuery, itemID, in.PetID, in.SourceType, in.SourceID, in.Title, in.Note, in.ScheduledFor, in.Status, in.CreatedByUserID, in.UpdatedByUserID)
		if err != nil {
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

func (r *ScheduledRepository) CancelHealthScheduledItem(ctx context.Context, in repo.CancelHealthScheduledItemInput) error {
	const query = `
		UPDATE scheduled_items
		SET
			status = 'CANCELLED',
			updated_at = NOW(),
			updated_by_user_id = $4,
			row_version = row_version + 1
		WHERE pet_id = $1
		  AND source_type = $2
		  AND source_id = $3
		  AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, in.PetID, in.SourceType, in.SourceID, in.UpdatedByUserID)
	return err
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
		WHERE pet_id = $1
		  AND source_type = $2
		  AND source_id = $3
		  AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, in.PetID, in.SourceType, in.SourceID, in.DeletedByUserID)
	return err
}

func (r *ScheduledRepository) CreateScheduledItemDispatch(ctx context.Context, in repo.CreateScheduledItemDispatchInput) error {
	const query = `
		INSERT INTO scheduled_item_push_dispatches (
			id,
			scheduled_item_id,
			dispatch_key,
			created_at
		) VALUES ($1,$2,$3,NOW())
	`
	_, err := r.db.Exec(ctx, query, in.ID, in.ScheduledItemID, in.DispatchKey)
	if err != nil {
		if isUniqueViolation(err) {
			return repo.ErrConflict
		}
		return err
	}
	return nil
}

func (r *ScheduledRepository) ListCalendarDayScheduledItems(ctx context.Context, petID uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItem, error) {
	return r.listCalendarDayScheduledItems(ctx, []uuid.UUID{petID}, dayStart, dayEnd)
}

func (r *ScheduledRepository) ListCalendarDayScheduledItemsForPets(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItem, error) {
	if len(petIDs) == 0 {
		return []model.ScheduledItem{}, nil
	}
	return r.listCalendarDayScheduledItems(ctx, petIDs, dayStart, dayEnd)
}

func (r *ScheduledRepository) listCalendarDayScheduledItems(ctx context.Context, petIDs []uuid.UUID, dayStart, dayEnd time.Time) ([]model.ScheduledItem, error) {
	const query = `
		SELECT
			id,
			pet_id,
			source_type,
			source_id,
			title,
			note,
			scheduled_for,
			recurrence_rule,
			recurrence_interval,
			recurrence_until,
			status,
			row_version,
			created_at,
			created_by_user_id,
			updated_at,
			updated_by_user_id,
			deleted_at,
			deleted_by_user_id
		FROM scheduled_items
		WHERE pet_id = ANY($1)
		  AND deleted_at IS NULL
		  AND status = 'ACTIVE'
		  AND scheduled_for >= $2
		  AND scheduled_for <= $3
		ORDER BY scheduled_for ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query, petIDs, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ScheduledItem, 0)
	for rows.Next() {
		var item model.ScheduledItem
		if err := rows.Scan(
			&item.ID,
			&item.PetID,
			&item.SourceType,
			&item.SourceID,
			&item.Title,
			&item.Note,
			&item.ScheduledFor,
			&item.RecurrenceRule,
			&item.RecurrenceInterval,
			&item.RecurrenceUntil,
			&item.Status,
			&item.RowVersion,
			&item.CreatedAt,
			&item.CreatedByUserID,
			&item.UpdatedAt,
			&item.UpdatedByUserID,
			&item.DeletedAt,
			&item.DeletedByUserID,
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
