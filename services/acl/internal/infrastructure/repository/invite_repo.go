package pgrepo

import (
	"acl/internal/repository"
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

func (r *InviteRepository) Create(ctx context.Context, in repository.InviteCreateInput) (*repository.InviteView, error) {
	policyRaw, err := json.Marshal(in.Policy)
	if err != nil {
		return nil, err
	}

	const insertQ = `
		INSERT INTO pet_invites (
			id, pet_id, created_by_user_id, status, token_hash,
			code, expires_at, role_id, policy, base_preset_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			NOW(), NOW()
		)
	`

	_, err = r.db.Exec(ctx, insertQ,
		in.ID,
		in.PetID,
		in.CreatedByUserID,
		in.Status,
		in.TokenHash,
		in.Code,
		in.ExpiresAt,
		in.RoleID,
		policyRaw,
		in.BasePresetID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repository.ErrConflict
		}
		return nil, err
	}

	return r.getByID(ctx, in.ID)
}

func (r *InviteRepository) ListActiveByPet(ctx context.Context, petID uuid.UUID) ([]repository.InviteView, error) {
	const query = `
		SELECT
			i.id, i.pet_id, i.status, i.code, i.expires_at,
			i.base_preset_id, i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
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

	items := make([]repository.InviteView, 0)
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

func (r *InviteRepository) GetActiveByTokenHash(ctx context.Context, tokenHash string) (*repository.InviteView, error) {
	const query = `
		SELECT
			i.id, i.pet_id, i.status, i.code, i.expires_at,
			i.base_preset_id, i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.status = 'ACTIVE'
		  AND i.expires_at > NOW()
		  AND i.token_hash = $1
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, tokenHash)
	invite, err := scanInviteRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return invite, nil
}

func (r *InviteRepository) getByID(ctx context.Context, id uuid.UUID) (*repository.InviteView, error) {
	const query = `
		SELECT
			i.id, i.pet_id, i.status, i.code, i.expires_at,
			i.base_preset_id, i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.id = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return scanInviteRow(row)
}

func (r *InviteRepository) AcceptByCode(ctx context.Context, code string, acceptedByUserID uuid.UUID) (*repository.MemberView, uuid.UUID, error) {
	return r.acceptTx(ctx, "code", code, acceptedByUserID)
}

func (r *InviteRepository) AcceptByTokenHash(ctx context.Context, tokenHash string, acceptedByUserID uuid.UUID) (*repository.MemberView, uuid.UUID, error) {
	return r.acceptTx(ctx, "token_hash", tokenHash, acceptedByUserID)
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
			return repository.ErrConflict
		}
		return repository.ErrNotFound
	}
	return nil
}

func (r *InviteRepository) acceptTx(ctx context.Context, key string, value string, acceptedByUserID uuid.UUID) (*repository.MemberView, uuid.UUID, error) {
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
		return nil, uuid.Nil, repository.ErrNotFound
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

func (r *InviteRepository) selectInviteForUpdate(ctx context.Context, tx pgx.Tx, key string, value string) (*repository.InviteView, error) {
	query := `
		SELECT
			i.id, i.pet_id, i.status, i.code, i.expires_at,
			i.base_preset_id, i.created_by_user_id, i.created_at,
			i.consumed_at, i.consumed_by_user_id, i.policy,
			r.id, r.kind, r.pet_id, COALESCE(r.code, ''), r.title, r.created_by_user_id
		FROM pet_invites i
		JOIN roles r ON r.id = i.role_id
		WHERE i.status = 'ACTIVE'
	`
	switch key {
	case "code":
		query += ` AND i.code = $1`
	case "token_hash":
		query += ` AND i.token_hash = $1`
	default:
		return nil, repository.ErrNotFound
	}
	query += ` FOR UPDATE`

	row := tx.QueryRow(ctx, query, value)
	invite, err := scanInviteRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return invite, nil
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

func (r *InviteRepository) upsertMembership(ctx context.Context, tx pgx.Tx, invite *repository.InviteView, acceptedByUserID uuid.UUID) (uuid.UUID, error) {
	policyRaw, err := json.Marshal(invite.Policy)
	if err != nil {
		return uuid.Nil, err
	}

	memberID := uuid.New()
	const query = `
		INSERT INTO pet_memberships (
			id, pet_id, user_id, status, role_id, policy, base_preset_id,
			is_primary_owner, created_by_user_id, created_at, updated_at,
			removed_at, removed_by_user_id
		) VALUES (
			$1, $2, $3, 'ACTIVE', $4, $5, $6,
			FALSE, $7, NOW(), NOW(),
			NULL, NULL
		)
		ON CONFLICT (pet_id, user_id) DO UPDATE SET
			status = 'ACTIVE',
			role_id = EXCLUDED.role_id,
			policy = EXCLUDED.policy,
			base_preset_id = EXCLUDED.base_preset_id,
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
		invite.BasePresetID,
		invite.CreatedByUserID,
	).Scan(&memberID)
	if err != nil {
		return uuid.Nil, err
	}
	return memberID, nil
}

func (r *InviteRepository) getMemberByID(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) (*repository.MemberView, error) {
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

type inviteScanner interface {
	Scan(dest ...any) error
}

func scanInviteRow(s inviteScanner) (*repository.InviteView, error) {
	var (
		inv       repository.InviteView
		role      repository.RoleView
		policyRaw []byte
	)

	err := s.Scan(
		&inv.ID,
		&inv.PetID,
		&inv.Status,
		&inv.Code,
		&inv.ExpiresAt,
		&inv.BasePresetID,
		&inv.CreatedByUserID,
		&inv.CreatedAt,
		&inv.ConsumedAt,
		&inv.ConsumedByUserID,
		&policyRaw,
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

	if err := json.Unmarshal(policyRaw, &inv.Policy); err != nil {
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

var _ repository.InviteRepository = (*InviteRepository)(nil)
