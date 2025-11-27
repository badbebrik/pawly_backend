package oauth_identity_repo

import (
	"auth/internal/model"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OAuthIdentityRepository interface {
	Create(ctx context.Context, identity *model.OAuthIdentity) error
	GetByProviderAndExternalID(ctx context.Context, provider, externalID string) (*model.OAuthIdentity, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthIdentity, error)
	GetByEmail(ctx context.Context, provider, email string) (*model.OAuthIdentity, error)
}

type OAuthIdentityRepo struct {
	db *pgxpool.Pool
}

func NewOAuthIdentityRepo(db *pgxpool.Pool) *OAuthIdentityRepo {
	return &OAuthIdentityRepo{db: db}
}

func (r *OAuthIdentityRepo) Create(ctx context.Context, identity *model.OAuthIdentity) error {
	query := `
        INSERT INTO oauth_identities (id, user_id, provider, external_id, email, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    `

	_, err := r.db.Exec(ctx, query,
		identity.ID,
		identity.UserID,
		identity.Provider,
		identity.ExternalID,
		identity.Email,
	)

	return err
}

func (r *OAuthIdentityRepo) GetByProviderAndExternalID(ctx context.Context, provider, externalID string) (*model.OAuthIdentity, error) {
	query := `
        SELECT id, user_id, provider, external_id, email, created_at, updated_at
        FROM oauth_identities
        WHERE provider = $1 AND external_id = $2
    `

	row := r.db.QueryRow(ctx, query, provider, externalID)

	var oi model.OAuthIdentity

	err := row.Scan(
		&oi.ID,
		&oi.UserID,
		&oi.Provider,
		&oi.ExternalID,
		&oi.Email,
		&oi.CreatedAt,
		&oi.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrIdentityNotFound
	}

	return &oi, err
}

func (r *OAuthIdentityRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.OAuthIdentity, error) {
	query := `
        SELECT id, user_id, provider, external_id, email, created_at, updated_at
        FROM oauth_identities
        WHERE user_id = $1
    `

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.OAuthIdentity

	for rows.Next() {
		var oi model.OAuthIdentity
		err := rows.Scan(
			&oi.ID,
			&oi.UserID,
			&oi.Provider,
			&oi.ExternalID,
			&oi.Email,
			&oi.CreatedAt,
			&oi.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, oi)
	}

	return result, nil
}

func (r *OAuthIdentityRepo) GetByEmail(ctx context.Context, provider, email string) (*model.OAuthIdentity, error) {
	query := `
        SELECT id, user_id, provider, external_id, email, created_at, updated_at
        FROM oauth_identities
        WHERE provider = $1 AND email = $2
    `

	row := r.db.QueryRow(ctx, query, provider, email)

	var oi model.OAuthIdentity

	err := row.Scan(
		&oi.ID,
		&oi.UserID,
		&oi.Provider,
		&oi.ExternalID,
		&oi.Email,
		&oi.CreatedAt,
		&oi.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrIdentityNotFound
	}

	return &oi, err
}
