package repository

import (
	"context"
	"errors"

	"push/internal/model"
	repo "push/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushRepository struct {
	db *pgxpool.Pool
}

func NewPushRepository(db *pgxpool.Pool) *PushRepository {
	return &PushRepository{db: db}
}

func (r *PushRepository) UpsertDeviceToken(ctx context.Context, in repo.UpsertDeviceTokenInput) (*model.DeviceToken, error) {
	const query = `
		INSERT INTO device_tokens (
			id, user_id, device_id, platform, push_token, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id, device_id)
		DO UPDATE SET
			platform = EXCLUDED.platform,
			push_token = EXCLUDED.push_token,
			updated_at = NOW()
		RETURNING id, user_id, device_id, platform, push_token, created_at, updated_at
	`
	var item model.DeviceToken
	err := r.db.QueryRow(ctx, query, in.ID, in.UserID, in.DeviceID, in.Platform, in.PushToken).Scan(
		&item.ID,
		&item.UserID,
		&item.DeviceID,
		&item.Platform,
		&item.PushToken,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}
	return &item, nil
}

func (r *PushRepository) DeleteDeviceToken(ctx context.Context, in repo.DeleteDeviceTokenInput) error {
	const query = `DELETE FROM device_tokens WHERE user_id = $1 AND device_id = $2`
	_, err := r.db.Exec(ctx, query, in.UserID, in.DeviceID)
	return err
}

func (r *PushRepository) ListDeviceTokensByUser(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	const query = `
		SELECT id, user_id, device_id, platform, push_token, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY created_at ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.DeviceToken, 0)
	for rows.Next() {
		var item model.DeviceToken
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.DeviceID,
			&item.Platform,
			&item.PushToken,
			&item.CreatedAt,
			&item.UpdatedAt,
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

func (r *PushRepository) GetPetPushSettings(ctx context.Context, userID, petID uuid.UUID) (*model.PetPushSettings, error) {
	const query = `
		SELECT user_id, pet_id, scheduled_items_enabled, created_at, updated_at
		FROM pet_push_settings
		WHERE user_id = $1 AND pet_id = $2
	`
	var item model.PetPushSettings
	err := r.db.QueryRow(ctx, query, userID, petID).Scan(
		&item.UserID,
		&item.PetID,
		&item.ScheduledItemsEnabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *PushRepository) UpsertPetPushSettings(ctx context.Context, in repo.UpsertPetPushSettingsInput) (*model.PetPushSettings, error) {
	const query = `
		INSERT INTO pet_push_settings (
			user_id, pet_id, scheduled_items_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (user_id, pet_id)
		DO UPDATE SET
			scheduled_items_enabled = EXCLUDED.scheduled_items_enabled,
			updated_at = NOW()
		RETURNING user_id, pet_id, scheduled_items_enabled, created_at, updated_at
	`
	var item model.PetPushSettings
	err := r.db.QueryRow(ctx, query, in.UserID, in.PetID, in.ScheduledItemsEnabled).Scan(
		&item.UserID,
		&item.PetID,
		&item.ScheduledItemsEnabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var _ repo.PushRepository = (*PushRepository)(nil)
