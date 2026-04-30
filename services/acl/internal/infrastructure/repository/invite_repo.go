package pgrepo

import (
	"acl/internal/application/ports"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InviteRepository struct {
	db *pgxpool.Pool
}

func NewInviteRepository(db *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(ctx context.Context, in ports.InviteCreateInput) (*ports.InviteView, error) {
	policyRaw, err := json.Marshal(in.Policy)
	if err != nil {
		return nil, err
	}

	const insertQ = `
		INSERT INTO pet_invites (
			id, pet_id, created_by_user_id, status, token,
			code, expires_at, role_id, policy,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			NOW(), NOW()
		)
	`

	_, err = r.db.Exec(ctx, insertQ,
		in.ID,
		in.PetID,
		in.CreatedByUserID,
		in.Status,
		in.Token,
		in.Code,
		in.ExpiresAt,
		in.RoleID,
		policyRaw,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ports.ErrConflict
		}
		return nil, err
	}

	return r.getByID(ctx, in.ID)
}

func (r *InviteRepository) ListActiveByPet(ctx context.Context, petID uuid.UUID) ([]ports.InviteView, error) {
	const query = `
		SELECT
			i.id, i.pet_id, i.status, i.token, i.code, i.expires_at,
			i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.policy, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.pet_id = $1
		  AND i.status = 'ACTIVE'
		  AND i.expires_at > NOW()
		ORDER BY i.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ports.InviteView, 0)
	for rows.Next() {
		inv, err := scanInviteRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *inv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *InviteRepository) GetActiveByToken(ctx context.Context, token string) (*ports.InviteView, error) {
	const query = `
		SELECT
			i.id, i.pet_id, i.status, i.token, i.code, i.expires_at,
			i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.policy, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.status = 'ACTIVE'
		  AND i.expires_at > NOW()
		  AND i.token = $1
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, token)
	invite, err := scanInviteRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return invite, nil
}

func (r *InviteRepository) getByID(ctx context.Context, id uuid.UUID) (*ports.InviteView, error) {
	const query = `
		SELECT
			i.id, i.pet_id, i.status, i.token, i.code, i.expires_at,
			i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.policy, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return scanInviteRow(row)
}

func (r *InviteRepository) AcceptByCode(ctx context.Context, code string, acceptedByUserID uuid.UUID) (*ports.MemberView, uuid.UUID, error) {
	return r.acceptTx(ctx, "code", code, acceptedByUserID)
}

func (r *InviteRepository) AcceptByToken(ctx context.Context, token string, acceptedByUserID uuid.UUID) (*ports.MemberView, uuid.UUID, error) {
	return r.acceptTx(ctx, "token", token, acceptedByUserID)
}

func (r *InviteRepository) RevokeByID(ctx context.Context, petID, inviteID uuid.UUID) error {
	const query = `
		UPDATE pet_invites
		SET status = 'REVOKED',
		    updated_at = NOW()
		WHERE id = $1
		  AND pet_id = $2
		  AND status = 'ACTIVE'
	`
	cmd, err := r.db.Exec(ctx, query, inviteID, petID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, err := r.existsByIDAndPet(ctx, inviteID, petID)
		if err != nil {
			return err
		}
		if exists {
			return ports.ErrConflict
		}
		return ports.ErrNotFound
	}
	return nil
}

func (r *InviteRepository) RotateTokenByID(ctx context.Context, petID, inviteID uuid.UUID, token string) (*ports.InviteView, error) {
	const updateQ = `
		UPDATE pet_invites
		SET token = $3,
		    updated_at = NOW()
		WHERE id = $1
		  AND pet_id = $2
		  AND status = 'ACTIVE'
		  AND expires_at > NOW()
		RETURNING id
	`

	var updatedID uuid.UUID
	err := r.db.QueryRow(ctx, updateQ, inviteID, petID, token).Scan(&updatedID)
	if err == nil {
		return r.getByID(ctx, updatedID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ports.ErrConflict
		}
		return nil, err
	}

	state, exists, stateErr := r.getInviteStateByIDAndPet(ctx, inviteID, petID)
	if stateErr != nil {
		return nil, stateErr
	}
	if !exists {
		return nil, ports.ErrNotFound
	}
	if state != "ACTIVE" {
		return nil, ports.ErrConflict
	}
	return nil, ports.ErrConflict
}

func (r *InviteRepository) acceptTx(ctx context.Context, key string, value string, acceptedByUserID uuid.UUID) (*ports.MemberView, uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	invite, err := r.selectInviteForUpdate(ctx, tx, key, value)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		if err := r.markInviteExpired(ctx, tx, invite.ID); err != nil {
			return nil, uuid.Nil, err
		}
		return nil, uuid.Nil, ports.ErrNotFound
	}
	if exists, err := r.activeMembershipExistsForUpdate(ctx, tx, invite.PetID, acceptedByUserID); err != nil {
		return nil, uuid.Nil, err
	} else if exists {
		return nil, uuid.Nil, ports.ErrConflict
	}

	if err := r.consumeInvite(ctx, tx, invite.ID, acceptedByUserID); err != nil {
		return nil, uuid.Nil, err
	}

	memberID, err := r.upsertMembership(ctx, tx, invite, acceptedByUserID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	member, err := r.getMemberByID(ctx, tx, memberID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, uuid.Nil, err
	}
	return member, invite.PetID, nil
}

func (r *InviteRepository) selectInviteForUpdate(ctx context.Context, tx pgx.Tx, key string, value string) (*ports.InviteView, error) {
	query := `
		SELECT
			i.id, i.pet_id, i.status, i.token, i.code, i.expires_at,
			i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.policy, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.status = 'ACTIVE'
	`
	switch key {
	case "code":
		query += ` AND i.code = $1`
	case "token":
		query += ` AND i.token = $1`
	default:
		return nil, ports.ErrNotFound
	}
	query += ` FOR UPDATE`

	row := tx.QueryRow(ctx, query, value)
	invite, err := scanInviteRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return invite, nil
}

func (r *InviteRepository) activeMembershipExistsForUpdate(ctx context.Context, tx pgx.Tx, petID, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT 1
		FROM pet_memberships
		WHERE pet_id = $1
		  AND user_id = $2
		  AND status = 'ACTIVE'
		FOR UPDATE
	`
	var x int
	err := tx.QueryRow(ctx, query, petID, userID).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *InviteRepository) markInviteExpired(ctx context.Context, tx pgx.Tx, inviteID uuid.UUID) error {
	const query = `
		UPDATE pet_invites
		SET status = 'EXPIRED', updated_at = NOW()
		WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, inviteID)
	return err
}

func (r *InviteRepository) consumeInvite(ctx context.Context, tx pgx.Tx, inviteID, acceptedByUserID uuid.UUID) error {
	const query = `
		UPDATE pet_invites
		SET status = 'CONSUMED',
		    consumed_at = NOW(),
		    consumed_by_user_id = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, inviteID, acceptedByUserID)
	return err
}

func (r *InviteRepository) upsertMembership(ctx context.Context, tx pgx.Tx, invite *ports.InviteView, acceptedByUserID uuid.UUID) (uuid.UUID, error) {
	policyRaw, err := json.Marshal(invite.Policy)
	if err != nil {
		return uuid.Nil, err
	}

	memberID := uuid.New()
	const query = `
		INSERT INTO pet_memberships (
			id, pet_id, user_id, status, role_id, policy,
			is_primary_owner, created_by_user_id, created_at, updated_at,
			removed_at, removed_by_user_id
		) VALUES (
			$1, $2, $3, 'ACTIVE', $4, $5,
			FALSE, $6, NOW(), NOW(),
			NULL, NULL
		)
		ON CONFLICT (pet_id, user_id) DO UPDATE SET
			status = 'ACTIVE',
			role_id = EXCLUDED.role_id,
			policy = EXCLUDED.policy,
			updated_at = NOW(),
			removed_at = NULL,
			removed_by_user_id = NULL
		RETURNING id
	`
	err = tx.QueryRow(ctx, query,
		memberID,
		invite.PetID,
		acceptedByUserID,
		invite.Role.ID,
		policyRaw,
		invite.CreatedByUserID,
	).Scan(&memberID)
	if err != nil {
		return uuid.Nil, err
	}
	return memberID, nil
}

func (r *InviteRepository) getMemberByID(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (*ports.MemberView, error) {
	const query = `
		SELECT
			m.id, m.pet_id, m.user_id, m.status, m.is_primary_owner, m.policy, m.created_at, m.updated_at,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.policy, r.created_by_user_id
		FROM pet_memberships m
		JOIN roles r ON r.id = m.role_id
		WHERE m.id = $1
	`
	row := tx.QueryRow(ctx, query, memberID)
	member, err := scanMemberRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return member, nil
}

type inviteScanner interface {
	Scan(dest ...any) error
}

func scanInviteRow(s inviteScanner) (*ports.InviteView, error) {
	var (
		inv           ports.InviteView
		role          ports.RoleView
		invPolicyRaw  []byte
		rolePolicyRaw []byte
	)

	err := s.Scan(
		&inv.ID,
		&inv.PetID,
		&inv.Status,
		&inv.Token,
		&inv.Code,
		&inv.ExpiresAt,
		&inv.CreatedByUserID,
		&inv.CreatedAt,
		&inv.ConsumedAt,
		&inv.ConsumedByUserID,
		&invPolicyRaw,
		&role.ID,
		&role.Kind,
		&role.PetID,
		&role.Code,
		&role.Title,
		&rolePolicyRaw,
		&role.CreatedByUserID,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(invPolicyRaw, &inv.Policy); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rolePolicyRaw, &role.Policy); err != nil {
		return nil, err
	}
	inv.Role = role
	return &inv, nil
}

func (r *InviteRepository) existsByIDAndPet(ctx context.Context, inviteID, petID uuid.UUID) (bool, error) {
	const query = `
		SELECT 1
		FROM pet_invites
		WHERE id = $1
		  AND pet_id = $2
		LIMIT 1
	`
	var x int
	err := r.db.QueryRow(ctx, query, inviteID, petID).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *InviteRepository) getInviteStateByIDAndPet(ctx context.Context, inviteID, petID uuid.UUID) (string, bool, error) {
	const query = `
		SELECT status
		FROM pet_invites
		WHERE id = $1
		  AND pet_id = $2
		LIMIT 1
	`
	var status string
	err := r.db.QueryRow(ctx, query, inviteID, petID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return status, true, nil
}

var _ ports.InviteRepository = (*InviteRepository)(nil)
