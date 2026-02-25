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

type PresetRepository struct {
	db *pgxpool.Pool
}

func NewPresetRepository(db *pgxpool.Pool) *PresetRepository {
	return &PresetRepository{db: db}
}

func (r *PresetRepository) ListSystem(ctx context.Context) ([]repository.PermissionPresetView, error) {
	const query = `
		SELECT id, name, COALESCE(role_code, ''), policy
		FROM permission_presets
		WHERE is_system = TRUE
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]repository.PermissionPresetView, 0)
	for rows.Next() {
		var (
			preset    repository.PermissionPresetView
			policyRaw []byte
		)
		if err := rows.Scan(&preset.ID, &preset.Name, &preset.RoleCode, &policyRaw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(policyRaw, &preset.Policy); err != nil {
			return nil, err
		}
		items = append(items, preset)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PresetRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM permission_presets WHERE id = $1 LIMIT 1`
	var x int
	err := r.db.QueryRow(ctx, query, id).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

var _ repository.PresetRepository = (*PresetRepository)(nil)
