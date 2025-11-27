package device_repo

import (
	"auth/internal/model"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserDeviceRepository interface {
	Upsert(ctx context.Context, d *model.Device) error
	GetByUserAndDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*model.Device, error)
	SetActive(ctx context.Context, userID uuid.UUID, deviceID string, active bool) error
	DeactivateByFCMToken(ctx context.Context, token string) error
	DeactivateAllByUserID(ctx context.Context, userID uuid.UUID) error
}

type UserDeviceRepo struct {
	db *pgxpool.Pool
}

func NewUserDeviceRepo(db *pgxpool.Pool) *UserDeviceRepo {
	return &UserDeviceRepo{db: db}
}

func (r *UserDeviceRepo) GetByUserAndDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*model.Device, error) {
	query := `
        SELECT id, user_id, device_id, platform, app_version, locale, fcm_token, is_active, created_at, updated_at
        FROM devices
        WHERE user_id = $1 AND device_id = $2
    `

	row := r.db.QueryRow(ctx, query, userID, deviceID)

	var d model.Device
	err := row.Scan(
		&d.ID,
		&d.UserID,
		&d.DeviceID,
		&d.Platform,
		&d.AppVersion,
		&d.Locale,
		&d.FCMToken,
		&d.IsActive,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}

	return &d, nil
}

func (r *UserDeviceRepo) Upsert(ctx context.Context, d *model.Device) error {

	if d.FCMToken != nil {
		var existingUserID uuid.UUID

		err := r.db.QueryRow(ctx,
			`SELECT user_id FROM devices WHERE fcm_token = $1`,
			*d.FCMToken,
		).Scan(&existingUserID)

		if err == nil && existingUserID != d.UserID {
			return ErrTokenInUse
		}

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	_, err := r.GetByUserAndDevice(ctx, d.UserID, d.DeviceID)

	if errors.Is(err, ErrDeviceNotFound) {
		query := `
            INSERT INTO devices 
            (id, user_id, device_id, platform, app_version, locale, fcm_token, is_active, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, NOW(), NOW())
        `

		if d.ID == uuid.Nil {
			d.ID = uuid.New()
		}

		_, err := r.db.Exec(ctx, query,
			d.ID,
			d.UserID,
			d.DeviceID,
			d.Platform,
			d.AppVersion,
			d.Locale,
			d.FCMToken,
		)
		return err
	}

	if err != nil {
		return err
	}

	query := `
        UPDATE devices
        SET platform = $3,
            app_version = $4,
            locale = $5,
            fcm_token = $6,
            is_active = TRUE,
            updated_at = NOW()
        WHERE user_id = $1 AND device_id = $2
    `

	_, err = r.db.Exec(ctx, query,
		d.UserID,
		d.DeviceID,
		d.Platform,
		d.AppVersion,
		d.Locale,
		d.FCMToken,
	)

	return err
}

func (r *UserDeviceRepo) SetActive(ctx context.Context, userID uuid.UUID, deviceID string, active bool) error {
	query := `
        UPDATE devices
        SET is_active = $3, updated_at = NOW()
        WHERE user_id = $1 AND device_id = $2
    `
	cmd, err := r.db.Exec(ctx, query, userID, deviceID, active)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrDeviceNotFound
	}
	return nil
}

func (r *UserDeviceRepo) DeactivateByFCMToken(ctx context.Context, token string) error {
	query := `
        UPDATE devices
        SET is_active = FALSE, updated_at = NOW()
        WHERE fcm_token = $1
    `
	_, err := r.db.Exec(ctx, query, token)
	return err
}

func (r *UserDeviceRepo) DeactivateAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
        UPDATE devices
        SET is_active = FALSE, updated_at = NOW()
        WHERE user_id = $1
    `
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
