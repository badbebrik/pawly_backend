package repository

import (
	"context"
	"errors"
	"pet/internal/model"
	repo "pet/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PetRepository struct {
	db *pgxpool.Pool
}

func NewPetRepository(db *pgxpool.Pool) *PetRepository {
	return &PetRepository{db: db}
}

func (r *PetRepository) Create(ctx context.Context, in repo.CreatePetInput) (*model.Pet, error) {
	p := in.Pet
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
		INSERT INTO pets (
			id, owner_user_id, name, species_id, sex, birth_date,
			breed_id, custom_breed_name,
			pattern_id, custom_pattern_name,
			is_neutered, is_outdoor, profile_photo_file_id,
			microchip_id, microchip_installed_at,
			status, missing_since, archived_at,
			created_at, updated_at, row_version
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,
			$9,$10,
			$11,$12,$13,
			$14,$15,
			$16,$17,$18,
			NOW(),NOW(),1
		)
	`
	_, err = tx.Exec(ctx, query,
		p.ID, p.OwnerUserID, p.Name, p.SpeciesID, p.Sex, p.BirthDate,
		p.BreedID, p.CustomBreedName,
		p.PatternID, p.CustomPatternName,
		p.IsNeutered, p.IsOutdoor, p.ProfilePhotoFileID,
		p.MicrochipID, p.MicrochipInstalledAt,
		p.Status, p.MissingSince, p.ArchivedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if err := syncPetColors(ctx, tx, p.ID, p.Colors); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, p.ID)
}

func (r *PetRepository) DeleteByID(ctx context.Context, petID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM pets WHERE id = $1`, petID)
	return err
}

func (r *PetRepository) GetByID(ctx context.Context, petID uuid.UUID) (*model.Pet, error) {
	const query = `
		SELECT
			id, owner_user_id, row_version, name, species_id, sex, birth_date,
			breed_id, custom_breed_name,
			pattern_id, custom_pattern_name,
			is_neutered, is_outdoor, profile_photo_file_id,
			microchip_id, microchip_installed_at,
			status, missing_since, archived_at,
			created_at, updated_at
		FROM pets
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, petID)
	pet, err := scanPet(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	colorsByPet, err := listColorsByPetIDs(ctx, r.db, []uuid.UUID{pet.ID})
	if err != nil {
		return nil, err
	}
	pet.Colors = colorsByPet[pet.ID]
	return pet, nil
}

func (r *PetRepository) ListSpecies(ctx context.Context) ([]model.Species, error) {
	const query = `
		SELECT id, code, name_ru, name_en, icon_key, sort_order, is_active, created_at, updated_at
		FROM species
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Species, 0)
	for rows.Next() {
		var item model.Species
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.NameRu,
			&item.NameEn,
			&item.IconKey,
			&item.SortOrder,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PetRepository) ListBreeds(ctx context.Context) ([]model.Breed, error) {
	const query = `
		SELECT id, species_id, name_ru, name_en, sort_order, is_active, created_at, updated_at
		FROM breeds
		WHERE is_active = TRUE
		ORDER BY species_id ASC, sort_order ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Breed, 0)
	for rows.Next() {
		var item model.Breed
		if err := rows.Scan(
			&item.ID,
			&item.SpeciesID,
			&item.NameRu,
			&item.NameEn,
			&item.SortOrder,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PetRepository) ListPatterns(ctx context.Context) ([]model.Pattern, error) {
	const query = `
		SELECT id, species_id, name_ru, name_en, icon_key, sort_order, is_active, created_at, updated_at
		FROM patterns
		WHERE is_active = TRUE
		ORDER BY species_id ASC NULLS FIRST, sort_order ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Pattern, 0)
	for rows.Next() {
		var item model.Pattern
		if err := rows.Scan(
			&item.ID,
			&item.SpeciesID,
			&item.NameRu,
			&item.NameEn,
			&item.IconKey,
			&item.SortOrder,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PetRepository) ListColorPresets(ctx context.Context) ([]model.ColorPreset, error) {
	const query = `
		SELECT id, name_ru, name_en, hex, sort_order, is_active, created_at, updated_at
		FROM color_presets
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, id ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ColorPreset, 0)
	for rows.Next() {
		var item model.ColorPreset
		if err := rows.Scan(
			&item.ID,
			&item.NameRu,
			&item.NameEn,
			&item.Hex,
			&item.SortOrder,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PetRepository) ListByIDs(ctx context.Context, ids []uuid.UUID, includeArchived bool, offset, limit int) ([]model.Pet, int, error) {
	if len(ids) == 0 {
		return []model.Pet{}, 0, nil
	}

	countQuery := `SELECT COUNT(1) FROM pets WHERE id = ANY($1::uuid[])`
	if !includeArchived {
		countQuery += ` AND status <> 'ARCHIVED'`
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, ids).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT
			id, owner_user_id, row_version, name, species_id, sex, birth_date,
			breed_id, custom_breed_name,
			pattern_id, custom_pattern_name,
			is_neutered, is_outdoor, profile_photo_file_id,
			microchip_id, microchip_installed_at,
			status, missing_since, archived_at,
			created_at, updated_at
		FROM pets
		WHERE id = ANY($1::uuid[])
	`
	if !includeArchived {
		listQuery += ` AND status <> 'ARCHIVED'`
	}
	listQuery += ` ORDER BY created_at DESC OFFSET $2 LIMIT $3`

	rows, err := r.db.Query(ctx, listQuery, ids, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Pet, 0)
	for rows.Next() {
		pet, err := scanPet(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *pet)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	colorsByPet, err := listColorsByPetIDs(ctx, r.db, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].Colors = colorsByPet[items[i].ID]
	}

	return items, total, nil
}

func (r *PetRepository) Update(ctx context.Context, petID uuid.UUID, rowVersion int, pet model.Pet) (*model.Pet, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
		UPDATE pets
		SET name = $3,
		    species_id = $4,
		    sex = $5,
		    birth_date = $6,
		    breed_id = $7,
		    custom_breed_name = $8,
		    pattern_id = $9,
		    custom_pattern_name = $10,
		    is_neutered = $11,
		    is_outdoor = $12,
		    profile_photo_file_id = $13,
		    microchip_id = $14,
		    microchip_installed_at = $15,
		    updated_at = NOW(),
		    row_version = row_version + 1
		WHERE id = $1
		  AND row_version = $2
	`

	cmd, err := tx.Exec(ctx, query,
		petID, rowVersion,
		pet.Name,
		pet.SpeciesID,
		pet.Sex,
		pet.BirthDate,
		pet.BreedID,
		pet.CustomBreedName,
		pet.PatternID,
		pet.CustomPatternName,
		pet.IsNeutered,
		pet.IsOutdoor,
		pet.ProfilePhotoFileID,
		pet.MicrochipID,
		pet.MicrochipInstalledAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if cmd.RowsAffected() == 0 {
		exists, err := r.existsByID(ctx, petID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		return nil, repo.ErrConflict
	}

	if err := syncPetColors(ctx, tx, petID, pet.Colors); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, petID)
}

func (r *PetRepository) UpdateOwner(ctx context.Context, petID uuid.UUID, rowVersion int, ownerUserID uuid.UUID) (*model.Pet, error) {
	const query = `
		UPDATE pets
		SET owner_user_id = $3,
		    updated_at = NOW(),
		    row_version = row_version + 1
		WHERE id = $1
		  AND row_version = $2
	`
	cmd, err := r.db.Exec(ctx, query, petID, rowVersion, ownerUserID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if cmd.RowsAffected() == 0 {
		exists, err := r.existsByID(ctx, petID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		return nil, repo.ErrConflict
	}

	return r.GetByID(ctx, petID)
}

func (r *PetRepository) UpdatePhoto(ctx context.Context, petID uuid.UUID, rowVersion int, fileID *uuid.UUID) (*model.Pet, error) {
	const query = `
		UPDATE pets
		SET profile_photo_file_id = $3,
		    updated_at = NOW(),
		    row_version = row_version + 1
		WHERE id = $1
		  AND row_version = $2
	`
	cmd, err := r.db.Exec(ctx, query, petID, rowVersion, fileID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if cmd.RowsAffected() == 0 {
		exists, err := r.existsByID(ctx, petID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		return nil, repo.ErrConflict
	}

	return r.GetByID(ctx, petID)
}

func (r *PetRepository) UpdateStatus(ctx context.Context, petID uuid.UUID, rowVersion int, status string, missingSince *time.Time, archivedAt *time.Time) (*model.Pet, error) {
	const query = `
		UPDATE pets
		SET status = $3,
		    missing_since = $4,
		    archived_at = $5,
		    updated_at = NOW(),
		    row_version = row_version + 1
		WHERE id = $1
		  AND row_version = $2
	`
	cmd, err := r.db.Exec(ctx, query, petID, rowVersion, status, missingSince, archivedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repo.ErrConflict
		}
		return nil, err
	}

	if cmd.RowsAffected() == 0 {
		exists, err := r.existsByID(ctx, petID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, repo.ErrNotFound
		}
		return nil, repo.ErrConflict
	}

	return r.GetByID(ctx, petID)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPet(s scanner) (*model.Pet, error) {
	var pet model.Pet

	err := s.Scan(
		&pet.ID, &pet.OwnerUserID, &pet.RowVersion, &pet.Name, &pet.SpeciesID, &pet.Sex, &pet.BirthDate,
		&pet.BreedID, &pet.CustomBreedName,
		&pet.PatternID, &pet.CustomPatternName,
		&pet.IsNeutered, &pet.IsOutdoor, &pet.ProfilePhotoFileID,
		&pet.MicrochipID, &pet.MicrochipInstalledAt,
		&pet.Status, &pet.MissingSince, &pet.ArchivedAt,
		&pet.CreatedAt, &pet.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	pet.Colors = []model.Color{}

	return &pet, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *PetRepository) existsByID(ctx context.Context, petID uuid.UUID) (bool, error) {
	const query = `SELECT 1 FROM pets WHERE id = $1 LIMIT 1`
	var x int
	err := r.db.QueryRow(ctx, query, petID).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type colorQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func listColorsByPetIDs(ctx context.Context, q colorQueryer, petIDs []uuid.UUID) (map[uuid.UUID][]model.Color, error) {
	result := make(map[uuid.UUID][]model.Color, len(petIDs))
	if len(petIDs) == 0 {
		return result, nil
	}

	const query = `
		SELECT id, pet_id, sort_order, preset_id, custom_name, custom_hex, created_at, updated_at
		FROM pet_colors
		WHERE pet_id = ANY($1::uuid[])
		ORDER BY pet_id ASC, sort_order ASC, id ASC
	`
	rows, err := q.Query(ctx, query, petIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var color model.Color
		if err := rows.Scan(
			&color.ID,
			&color.PetID,
			&color.SortOrder,
			&color.PresetID,
			&color.CustomName,
			&color.CustomHex,
			&color.CreatedAt,
			&color.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result[color.PetID] = append(result[color.PetID], color)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range petIDs {
		if _, ok := result[petIDs[i]]; !ok {
			result[petIDs[i]] = []model.Color{}
		}
	}

	return result, nil
}

func syncPetColors(ctx context.Context, q colorQueryer, petID uuid.UUID, colors []model.Color) error {
	if _, err := q.Exec(ctx, `DELETE FROM pet_colors WHERE pet_id = $1`, petID); err != nil {
		return err
	}
	if len(colors) == 0 {
		return nil
	}

	const query = `
		INSERT INTO pet_colors (
			id, pet_id, sort_order, preset_id, custom_name, custom_hex, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		)
	`
	for i := range colors {
		colorID := colors[i].ID
		if colorID == uuid.Nil {
			colorID = uuid.New()
		}
		if _, err := q.Exec(ctx, query,
			colorID,
			petID,
			colors[i].SortOrder,
			colors[i].PresetID,
			colors[i].CustomName,
			colors[i].CustomHex,
		); err != nil {
			if isUniqueViolation(err) {
				return repo.ErrConflict
			}
			return err
		}
	}
	return nil
}

var _ repo.PetRepository = (*PetRepository)(nil)
