package pgrepo

import (
	"acl/internal/repository"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) ListSystemAndPetRoles(ctx context.Context, petID uuid.UUID) ([]repository.RoleView, error) {
	const query = `
		SELECT id, kind, COALESCE(code, ''), title
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

	items := make([]repository.RoleView, 0)
	for rows.Next() {
		var role repository.RoleView
		if err := rows.Scan(&role.ID, &role.Kind, &role.Code, &role.Title); err != nil {
			return nil, err
		}
		items = append(items, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

var _ repository.RoleRepository = (*RoleRepository)(nil)
