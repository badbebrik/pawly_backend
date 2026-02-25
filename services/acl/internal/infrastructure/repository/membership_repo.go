package pgrepo

import (
	"acl/internal/repository"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MembershipRepository struct {
	db *pgxpool.Pool
}

func NewMembershipRepository(db *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{db: db}
}

func (r *MembershipRepository) GetByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*repository.MembershipAccess, error) {
	const query = `
		SELECT id, status, is_primary_owner, policy
		FROM pet_memberships
		WHERE pet_id = $1 AND user_id = $2
	`

	var (
		access    repository.MembershipAccess
		policyRaw []byte
	)

	err := r.db.QueryRow(ctx, query, petID, userID).Scan(
		&access.MemberID,
		&access.Status,
		&access.IsPrimaryOwner,
		&policyRaw,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(policyRaw, &access.Policy); err != nil {
		return nil, err
	}

	return &access, nil
}

func (r *MembershipRepository) GetActiveByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*repository.MembershipAccess, error) {
	const query = `
		SELECT id, status, is_primary_owner, policy
		FROM pet_memberships
		WHERE pet_id = $1 AND user_id = $2 AND status = 'ACTIVE'
	`

	var (
		access    repository.MembershipAccess
		policyRaw []byte
	)

	err := r.db.QueryRow(ctx, query, petID, userID).Scan(
		&access.MemberID,
		&access.Status,
		&access.IsPrimaryOwner,
		&policyRaw,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(policyRaw, &access.Policy); err != nil {
		return nil, err
	}

	return &access, nil
}

func (r *MembershipRepository) ListActivePetIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT pet_id
		FROM pet_memberships
		WHERE user_id = $1 AND status = 'ACTIVE'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	petIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		petIDs = append(petIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return petIDs, nil
}

var _ repository.MembershipRepository = (*MembershipRepository)(nil)
