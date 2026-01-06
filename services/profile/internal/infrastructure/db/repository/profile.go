package pgrepo

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"profile/internal/model"
	"profile/internal/repository"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	const query = `
        SELECT user_id, first_name, last_name, avatar_url, phone,
               locale, timezone, date_format, notifications,
               created_at, updated_at
        FROM profiles
        WHERE user_id = $1
    `

	row := r.db.QueryRow(ctx, query, userID)

	var p model.Profile
	var notifBytes []byte

	err := row.Scan(
		&p.UserID,
		&p.FirstName,
		&p.LastName,
		&p.AvatarURL,
		&p.Phone,
		&p.Locale,
		&p.Timezone,
		&p.DateFormat,
		&notifBytes,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if len(notifBytes) > 0 {
		var m map[string]any
		if err := json.Unmarshal(notifBytes, &m); err == nil {
			p.Notifications = m
		}
	} else {
		p.Notifications = map[string]any{}
	}

	return &p, nil
}

func (r *ProfileRepository) Create(ctx context.Context, p *model.Profile) error {
	if p.Notifications == nil {
		p.Notifications = map[string]any{}
	}
	notifJSON, err := json.Marshal(p.Notifications)
	if err != nil {
		return err
	}

	const query = `
        INSERT INTO profiles (
            user_id, first_name, last_name, avatar_url, phone,
            locale, timezone, date_format, notifications,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5,
            $6, $7, $8, $9,
            NOW(), NOW()
        )
    `

	_, err = r.db.Exec(ctx, query,
		p.UserID,
		p.FirstName,
		p.LastName,
		p.AvatarURL,
		p.Phone,
		p.Locale,
		p.Timezone,
		p.DateFormat,
		notifJSON,
	)
	return err
}

func (r *ProfileRepository) Update(ctx context.Context, p *model.Profile) error {
	if p.Notifications == nil {
		p.Notifications = map[string]any{}
	}
	notifJSON, err := json.Marshal(p.Notifications)
	if err != nil {
		return err
	}

	const query = `
        UPDATE profiles
        SET first_name    = $2,
            last_name     = $3,
            avatar_url    = $4,
            phone         = $5,
            locale        = $6,
            timezone      = $7,
            date_format   = $8,
            notifications = $9,
            updated_at    = NOW()
        WHERE user_id = $1
    `

	cmd, err := r.db.Exec(ctx, query,
		p.UserID,
		p.FirstName,
		p.LastName,
		p.AvatarURL,
		p.Phone,
		p.Locale,
		p.Timezone,
		p.DateFormat,
		notifJSON,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
