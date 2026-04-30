package pgrepo

import (
	"acl/internal/application/ports"
	"acl/internal/domain/model"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*ports.RoleView, error) {
	const query = `
		SELECT id, kind, pet_id, COALESCE(code, ''), title, policy, created_by_user_id
		FROM roles
		WHERE id = $1
	`

	var role ports.RoleView
	var policyRaw []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&role.ID,
		&role.Kind,
		&role.PetID,
		&role.Code,
		&role.Title,
		&policyRaw,
		&role.CreatedByUserID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(policyRaw, &role.Policy); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) ListSystemAndPetRoles(ctx context.Context, petID uuid.UUID) ([]ports.RoleView, error) {
	const query = `
		SELECT id, kind, pet_id, COALESCE(code, ''), title, policy, created_by_user_id
		FROM roles
		WHERE kind = 'SYSTEM'
		   OR (kind = 'CUSTOM' AND pet_id = $1)
		ORDER BY kind ASC, title ASC
	`

	rows, err := r.db.Query(ctx, query, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ports.RoleView, 0)
	for rows.Next() {
		var role ports.RoleView
		var policyRaw []byte
		if err := rows.Scan(&role.ID, &role.Kind, &role.PetID, &role.Code, &role.Title, &policyRaw, &role.CreatedByUserID); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(policyRaw, &role.Policy); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *RoleRepository) CreateCustom(ctx context.Context, petID uuid.UUID, title string, policy model.Policy, createdByUserID uuid.UUID) (*ports.RoleView, error) {
	policyRaw, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	roleID := uuid.New()
	const query = `
		INSERT INTO roles (
			id, kind, pet_id, code, title, policy, created_by_user_id, created_at, updated_at
		) VALUES (
			$1, 'CUSTOM', $2, NULL, $3, $4, $5, NOW(), NOW()
		)
	`
	_, err = r.db.Exec(ctx, query, roleID, petID, title, policyRaw, createdByUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrConflict
		}
		return nil, err
	}
	return r.GetByID(ctx, roleID)
}

func (r *RoleRepository) UpdateCustom(ctx context.Context, petID, roleID uuid.UUID, title string, policy model.Policy) (*ports.RoleView, error) {
	policyRaw, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	const query = `
		UPDATE roles
		SET title = $3,
		    policy = $4,
		    updated_at = NOW()
		WHERE id = $1
		  AND pet_id = $2
		  AND kind = 'CUSTOM'
	`
	cmd, err := r.db.Exec(ctx, query, roleID, petID, title, policyRaw)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrConflict
		}
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, ports.ErrNotFound
	}
	return r.GetByID(ctx, roleID)
}

func (r *RoleRepository) DeleteCustomIfUnused(ctx context.Context, petID, roleID uuid.UUID) error {
	const query = `
		WITH role_target AS (
			SELECT id
			FROM roles
			WHERE id = $1
			  AND kind = 'CUSTOM'
			  AND pet_id = $2
		),
		role_usage AS (
			SELECT EXISTS(
				SELECT 1
				FROM pet_memberships m
				WHERE m.role_id = $1
				  AND m.status = 'ACTIVE'
			) AS in_memberships,
			EXISTS(
				SELECT 1
				FROM pet_invites i
				WHERE i.role_id = $1
				  AND i.status = 'ACTIVE'
				  AND i.expires_at > NOW()
			) AS in_invites
		)
		DELETE FROM roles
		WHERE id IN (SELECT id FROM role_target)
		  AND NOT (SELECT in_memberships OR in_invites FROM role_usage)
	`
	cmd, err := r.db.Exec(ctx, query, roleID, petID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		exists, err := r.customRoleExists(ctx, petID, roleID)
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

func (r *RoleRepository) customRoleExists(ctx context.Context, petID, roleID uuid.UUID) (bool, error) {
	const query = `
		SELECT 1
		FROM roles
		WHERE id = $1
		  AND kind = 'CUSTOM'
		  AND pet_id = $2
		LIMIT 1
	`
	var x int
	err := r.db.QueryRow(ctx, query, roleID, petID).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var _ ports.RoleRepository = (*RoleRepository)(nil)
