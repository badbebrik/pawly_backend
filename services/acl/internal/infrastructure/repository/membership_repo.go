package pgrepo

import (
	"acl/internal/model"
	"acl/internal/repository"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (r *MembershipRepository) CreateOwner(ctx context.Context, petID, ownerUserID uuid.UUID, policy model.Policy) (*repository.MemberView, error) {
	policyRaw, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	memberID := uuid.New()
	const query = `
		WITH owner_role AS (
			SELECT id
			FROM roles
			WHERE kind = 'SYSTEM' AND code = 'OWNER'
			LIMIT 1
		),
		owner_preset AS (
			SELECT id
			FROM permission_presets
			WHERE is_system = TRUE AND role_code = 'OWNER'
			ORDER BY created_at ASC
			LIMIT 1
		)
		INSERT INTO pet_memberships (
			id, pet_id, user_id, status, role_id, policy, base_preset_id,
			is_primary_owner, created_by_user_id, created_at, updated_at, removed_at, removed_by_user_id
		)
		SELECT
			$1, $2, $3, 'ACTIVE', owner_role.id, $4, owner_preset.id,
			TRUE, $3, NOW(), NOW(), NULL, NULL
		FROM owner_role
		LEFT JOIN owner_preset ON TRUE
		ON CONFLICT (pet_id, user_id) DO UPDATE
		SET status = 'ACTIVE',
		    role_id = EXCLUDED.role_id,
		    policy = EXCLUDED.policy,
		    base_preset_id = EXCLUDED.base_preset_id,
		    is_primary_owner = TRUE,
		    updated_at = NOW(),
		    removed_at = NULL,
		    removed_by_user_id = NULL
		WHERE pet_memberships.is_primary_owner = TRUE
		RETURNING id
	`

	if err := r.db.QueryRow(ctx, query, memberID, petID, ownerUserID, policyRaw).Scan(&memberID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repository.ErrConflict
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrConflict
		}
		return nil, err
	}

	member, err := r.getActiveByIDAndPet(ctx, memberID, petID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, repository.ErrConflict
		}
		return nil, err
	}

	if member.Role.Code == "" {
		return nil, repository.ErrNotFound
	}

	return member, nil
}

func (r *MembershipRepository) GetActiveViewByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
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
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
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

func (r *MembershipRepository) ListActiveViewsByUser(ctx context.Context, userID uuid.UUID) ([]repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.user_id = $1
		  AND m.status = 'ACTIVE'
		ORDER BY m.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
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

func (r *MembershipRepository) GetByIDAndPet(ctx context.Context, petID, memberID uuid.UUID) (*repository.MemberView, error) {
	return r.getAnyByIDAndPet(ctx, memberID, petID)
}

func (r *MembershipRepository) TransferOwnership(ctx context.Context, petID, requesterUserID, targetMemberID uuid.UUID) (*repository.TransferOwnershipView, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	currentOwner, err := getActivePrimaryOwnerByPetAndUserTx(ctx, tx, petID, requesterUserID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, repository.ErrForbidden
		}
		return nil, err
	}

	targetMember, err := getMemberByIDAndPetTx(ctx, tx, targetMemberID, petID)
	if err != nil {
		return nil, err
	}
	if targetMember.Status != membershipStatusActive {
		return nil, repository.ErrConflict
	}
	if targetMember.ID == currentOwner.ID || targetMember.UserID == requesterUserID {
		return nil, repository.ErrConflict
	}

	ownerRoleID, ownerPresetID, ownerPolicy, err := getSystemRoleAndPresetByCodeTx(ctx, tx, "OWNER")
	if err != nil {
		return nil, err
	}
	coOwnerRoleID, coOwnerPresetID, coOwnerPolicy, err := getSystemRoleAndPresetByCodeTx(ctx, tx, "CO_OWNER")
	if err != nil {
		return nil, err
	}

	if err := updateMembershipOwnershipTx(ctx, tx, currentOwner.ID, false, coOwnerRoleID, coOwnerPresetID, coOwnerPolicy); err != nil {
		return nil, err
	}
	if err := updateMembershipOwnershipTx(ctx, tx, targetMember.ID, true, ownerRoleID, ownerPresetID, ownerPolicy); err != nil {
		return nil, err
	}

	previousOwner, err := getMemberByIDTx(ctx, tx, currentOwner.ID)
	if err != nil {
		return nil, err
	}
	newOwner, err := getMemberByIDTx(ctx, tx, targetMember.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &repository.TransferOwnershipView{
		PreviousOwner: *previousOwner,
		CurrentOwner:  *newOwner,
	}, nil
}

func (r *MembershipRepository) UpdatePermissions(
	ctx context.Context,
	petID, memberID uuid.UUID,
	roleID uuid.UUID,
	policy model.Policy,
	basePresetID *uuid.UUID,
) (*repository.MemberView, error) {
	policyRaw, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE pet_memberships
		SET role_id = $3,
		    policy = $4,
		    base_preset_id = $5,
		    updated_at = NOW()
		WHERE id = $1
		  AND pet_id = $2
		  AND status = 'ACTIVE'
	`
	cmd, err := r.db.Exec(ctx, query, memberID, petID, roleID, policyRaw, basePresetID)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, repository.ErrNotFound
	}

	return r.getActiveByIDAndPet(ctx, memberID, petID)
}

func (r *MembershipRepository) RemoveMember(ctx context.Context, petID, memberID, removedByUserID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		UPDATE pet_memberships
		SET status = 'REMOVED',
		    removed_at = NOW(),
		    removed_by_user_id = $3,
		    updated_at = NOW()
		WHERE id = $1
		  AND pet_id = $2
		  AND status = 'ACTIVE'
		  AND is_primary_owner = FALSE
	`
	cmd, err := r.db.Exec(ctx, query, memberID, petID, removedByUserID)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		existing, err := r.getAnyByIDAndPet(ctx, memberID, petID)
		if err != nil {
			if err == repository.ErrNotFound {
				return nil, repository.ErrNotFound
			}
			return nil, err
		}
		if existing.IsPrimaryOwner && existing.Status == "ACTIVE" {
			return nil, repository.ErrConflict
		}
		return nil, repository.ErrNotFound
	}

	return r.getAnyByIDAndPet(ctx, memberID, petID)
}

func (r *MembershipRepository) getActiveByIDAndPet(ctx context.Context, memberID, petID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.id = $1
		  AND m.pet_id = $2
		  AND m.status = 'ACTIVE'
	`
	return r.scanMemberView(ctx, query, memberID, petID)
}

func (r *MembershipRepository) getAnyByIDAndPet(ctx context.Context, memberID, petID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.id = $1
		  AND m.pet_id = $2
	`
	return r.scanMemberView(ctx, query, memberID, petID)
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

const membershipStatusActive = "ACTIVE"

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
		&role.PetID,
		&role.Code,
		&role.Title,
		&role.CreatedByUserID,
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

func getActivePrimaryOwnerByPetAndUserTx(ctx context.Context, tx pgx.Tx, petID, userID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.pet_id = $1
		  AND m.user_id = $2
		  AND m.status = 'ACTIVE'
		  AND m.is_primary_owner = TRUE
		FOR UPDATE
	`
	row := tx.QueryRow(ctx, query, petID, userID)
	member, err := scanMemberRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return member, nil
}

func getMemberByIDAndPetTx(ctx context.Context, tx pgx.Tx, memberID, petID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.id = $1
		  AND m.pet_id = $2
		FOR UPDATE
	`
	row := tx.QueryRow(ctx, query, memberID, petID)
	member, err := scanMemberRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return member, nil
}

func getMemberByIDTx(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (*repository.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.id = $1
	`
	row := tx.QueryRow(ctx, query, memberID)
	member, err := scanMemberRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return member, nil
}

func getSystemRoleAndPresetByCodeTx(ctx context.Context, tx pgx.Tx, code string) (uuid.UUID, *uuid.UUID, model.Policy, error) {
	const query = `
		SELECT
			r.id,
			p.id,
			p.policy
		FROM roles r
		LEFT JOIN permission_presets p
		  ON p.is_system = TRUE
		 AND p.role_code = r.code
		WHERE r.kind = 'SYSTEM'
		  AND r.code = $1
		ORDER BY p.created_at ASC NULLS LAST
		LIMIT 1
	`

	var (
		roleID    uuid.UUID
		presetID  *uuid.UUID
		policyRaw []byte
		policy    model.Policy
	)

	err := tx.QueryRow(ctx, query, code).Scan(&roleID, &presetID, &policyRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil, model.Policy{}, repository.ErrNotFound
		}
		return uuid.Nil, nil, model.Policy{}, err
	}
	if presetID == nil || len(policyRaw) == 0 {
		return uuid.Nil, nil, model.Policy{}, repository.ErrNotFound
	}
	if len(policyRaw) > 0 {
		if err := json.Unmarshal(policyRaw, &policy); err != nil {
			return uuid.Nil, nil, model.Policy{}, err
		}
	}

	return roleID, presetID, policy, nil
}

func updateMembershipOwnershipTx(
	ctx context.Context,
	tx pgx.Tx,
	memberID uuid.UUID,
	isPrimaryOwner bool,
	roleID uuid.UUID,
	basePresetID *uuid.UUID,
	policy model.Policy,
) error {
	policyRaw, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	const query = `
		UPDATE pet_memberships
		SET is_primary_owner = $2,
		    role_id = $3,
		    base_preset_id = $4,
		    policy = $5,
		    updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := tx.Exec(ctx, query, memberID, isPrimaryOwner, roleID, basePresetID, policyRaw)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

var _ repository.MembershipRepository = (*MembershipRepository)(nil)
