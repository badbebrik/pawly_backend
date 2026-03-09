package pgrepo

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
        SELECT user_id, first_name, last_name, phone, avatar_file_id,
               locale, timezone, date_format,
               public_contact_settings, extra_contacts,
               created_at, updated_at
        FROM profiles
        WHERE user_id = $1
    `

	row := r.db.QueryRow(ctx, query, userID)
	p, err := scanProfileRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return p, nil
}

func (r *ProfileRepository) GetByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]model.Profile, error) {
	if len(userIDs) == 0 {
		return []model.Profile{}, nil
	}

	const query = `
        SELECT user_id, first_name, last_name, phone, avatar_file_id,
               locale, timezone, date_format,
               public_contact_settings, extra_contacts,
               created_at, updated_at
        FROM profiles
        WHERE user_id = ANY($1)
    `

	rows, err := r.db.Query(ctx, query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Profile, 0, len(userIDs))
	for rows.Next() {
		profile, err := scanProfileRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *profile)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type profileScanner interface {
	Scan(dest ...any) error
}

func scanProfileRow(s profileScanner) (*model.Profile, error) {
	var p model.Profile
	var publicBytes []byte
	var extraBytes []byte

	if err := s.Scan(
		&p.UserID,
		&p.FirstName,
		&p.LastName,
		&p.Phone,
		&p.AvatarFileID,
		&p.Locale,
		&p.Timezone,
		&p.DateFormat,
		&publicBytes,
		&extraBytes,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if len(publicBytes) > 0 {
		_ = json.Unmarshal(publicBytes, &p.PublicContact)
	}
	if len(extraBytes) > 0 {
		_ = json.Unmarshal(extraBytes, &p.ExtraContacts)
	}
	if p.ExtraContacts == nil {
		p.ExtraContacts = model.ExtraContacts{}
	}

	return &p, nil
}

func (r *ProfileRepository) Create(ctx context.Context, p *model.Profile) error {
	publicJSON, err := json.Marshal(p.PublicContact)
	if err != nil {
		return err
	}
	extraJSON, err := json.Marshal(p.ExtraContacts)
	if err != nil {
		return err
	}

	const query = `
        INSERT INTO profiles (
            user_id, first_name, last_name, phone, avatar_file_id,
            locale, timezone, date_format,
            public_contact_settings, extra_contacts,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5,
            $6, $7, $8,
            $9, $10,
            NOW(), NOW()
        )
    `

	_, err = r.db.Exec(ctx, query,
		p.UserID,
		p.FirstName,
		p.LastName,
		p.Phone,
		p.AvatarFileID,
		p.Locale,
		p.Timezone,
		p.DateFormat,
		publicJSON,
		extraJSON,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrConflict
		}
	}
	return err
}

func (r *ProfileRepository) Update(ctx context.Context, p *model.Profile) error {
	publicJSON, err := json.Marshal(p.PublicContact)
	if err != nil {
		return err
	}
	extraJSON, err := json.Marshal(p.ExtraContacts)
	if err != nil {
		return err
	}

	const query = `
        UPDATE profiles
        SET first_name              = $2,
            last_name               = $3,
            phone                   = $4,
            avatar_file_id          = $5,
            locale                  = $6,
            timezone                = $7,
            date_format             = $8,
            public_contact_settings = $9,
            extra_contacts          = $10,
            updated_at              = NOW()
        WHERE user_id = $1
    `

	cmd, err := r.db.Exec(ctx, query,
		p.UserID,
		p.FirstName,
		p.LastName,
		p.Phone,
		p.AvatarFileID,
		p.Locale,
		p.Timezone,
		p.DateFormat,
		publicJSON,
		extraJSON,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
