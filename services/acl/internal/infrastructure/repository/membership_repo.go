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

func (r *MembershipRepository) GetActiveViewByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, COALESCE(r.code, ''), r.title
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.pet_id = $1
		  AND m.user_id = $2
		  AND m.status = 'ACTIVE'
	`

	return r.scanMemberView(ctx, query, petID, userID)
}

func (r *MembershipRepository) ListActiveViewsByPet(ctx context.Context, petID uuid.UUID) ([]repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, COALESCE(r.code, ''), r.title
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.pet_id = $1
		  AND m.status = 'ACTIVE'
		ORDER BY m.created_at ASC
	`

	rows, err := r.db.Query(ctx, query, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repository.MemberView, 0)
	for rows.Next() {
		member, err := scanMemberRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MembershipRepository) scanMemberView(ctx context.Context, query string, args ...any) (*repository.MemberView, error) {
	row := r.db.QueryRow(ctx, query, args...)
	member, err := scanMemberRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return member, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanMemberRow(s scanner) (*repository.MemberView, error) {
	var (
		member    repository.MemberView
		role      repository.RoleView
		policyRaw []byte
	)
	err := s.Scan(
		&member.ID,
		&member.PetID,
		&member.UserID,
		&member.Status,
		&member.IsPrimaryOwner,
		&policyRaw,
		&member.CreatedAt,
		&member.UpdatedAt,
		&role.ID,
		&role.Kind,
		&role.Code,
		&role.Title,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(policyRaw, &member.Policy); err != nil {
		return nil, err
	}
	member.Role = role
	return &member, nil
}

var _ repository.MembershipRepository = (*MembershipRepository)(nil)
