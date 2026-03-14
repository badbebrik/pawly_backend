package pgrepo

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (ur *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, is_verified, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    `
	_, err := ur.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.IsVerified,
		user.IsActive)

	if err != nil {
		if isUniqueViolation(err) {
			return ports.ErrConflict
		}
		return err
	}

	return nil
}

func (ur *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
        SELECT id, email, password_hash, is_verified, is_active, created_at, updated_at, last_logged_at, deleted_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `
	row := ur.db.QueryRow(ctx, query, id)
	var u model.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsVerified,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoggedAt,
		&u.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrNotFound
	}

	return &u, nil
}

func (ur *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
        SELECT id, email, password_hash, is_verified, is_active, created_at, updated_at, last_logged_at, deleted_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

	row := ur.db.QueryRow(ctx, query, email)

	var u model.User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsVerified,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoggedAt,
		&u.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ports.ErrNotFound
	}

	return &u, err
}

func (ur *UserRepo) SetVerified(ctx context.Context, id uuid.UUID) error {
	query := `
        UPDATE users 
        SET is_verified = TRUE, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `

	cmd, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (ur *UserRepo) UpdatePasswordHash(ctx context.Context, id uuid.UUID, newHash string) error {
	query := `
        UPDATE users
        SET password_hash = $2, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	cmd, err := ur.db.Exec(ctx, query, id, newHash)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}

	return nil
}

func (ur *UserRepo) UpdateEmail(ctx context.Context, id uuid.UUID, newEmail string) error {
	query := `
        UPDATE users
        SET email = $2, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	_, err := ur.db.Exec(ctx, query, id, newEmail)
	if err != nil {
		if isUniqueViolation(err) {
			return ports.ErrConflict
		}
		return err
	}
	return nil
}

func (ur *UserRepo) SetActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	query := `
        UPDATE users
        SET is_active = $2, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	cmd, err := ur.db.Exec(ctx, query, id, isActive)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (ur *UserRepo) UpdateLastLoginAt(ctx context.Context, id uuid.UUID, ts time.Time) error {
	query := `
        UPDATE users
        SET last_logged_at = $2, updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	cmd, err := ur.db.Exec(ctx, query, id, ts)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (ur *UserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
        UPDATE users
        SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
	cmd, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (ur *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	cmd, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}
