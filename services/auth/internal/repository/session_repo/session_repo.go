package session_repo

import (
	"auth/internal/model"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	GetActiveByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, newHash string, newExpiresAt time.Time) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}

type SessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepo(db *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, s *model.Session) error {
	query := `
        INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at, is_revoked, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, FALSE, NOW(), NOW())
    `
	_, err := r.db.Exec(ctx, query,
		s.ID,
		s.UserID,
		s.RefreshTokenHash,
		s.ExpiresAt,
	)

	return err
}

func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	query := `
        SELECT id, user_id, refresh_token_hash, expires_at, is_revoked, created_at, updated_at
        FROM sessions
        WHERE id = $1
    `

	row := r.db.QueryRow(ctx, query, id)

	var s model.Session

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.ExpiresAt,
		&s.IsRevoked,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &s, nil
}

func (r *SessionRepo) GetActiveByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	s, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.IsRevoked {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(s.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return s, nil
}

func (r *SessionRepo) UpdateRefreshToken(ctx context.Context, id uuid.UUID, newHash string, newExpires time.Time) error {
	query := `
        UPDATE sessions
        SET refresh_token_hash = $2, expires_at = $3, updated_at = NOW()
        WHERE id = $1 AND is_revoked = FALSE
    `

	cmd, err := r.db.Exec(ctx, query, id, newHash, newExpires)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
        UPDATE sessions
        SET is_revoked = TRUE, updated_at = NOW()
        WHERE id = $1 AND is_revoked = FALSE
    `

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepo) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	query := `
        UPDATE sessions
        SET is_revoked = TRUE, updated_at = NOW()
        WHERE user_id = $1 AND is_revoked = FALSE
    `
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
