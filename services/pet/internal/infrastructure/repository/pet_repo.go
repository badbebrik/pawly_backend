package repository

import (
	"context"
	"encoding/json"
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
	colorsRaw, err := json.Marshal(p.Colors)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO pets (
			id, owner_user_id, name, species_id, sex, birth_date,
			breed_source, system_breed_id, custom_breed_name,
			colors,
			coat_pattern_source, system_coat_pattern_id, custom_coat_pattern_name,
			is_neutered, is_outdoor, profile_photo_file_id,
			microchip_id, microchip_installed_at,
			status, missing_since, archived_at,
			created_at, updated_at, row_version
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,
			$10,
			$11,$12,$13,
			$14,$15,$16,
			$17,$18,
			$19,$20,$21,
			NOW(),NOW(),1
		)
	`
	_, err = r.db.Exec(ctx, query,
		p.ID, p.OwnerUserID, p.Name, p.SpeciesID, p.Sex, p.BirthDate,
		p.Breed.Source, p.Breed.SystemBreedID, p.Breed.CustomBreedName,
		colorsRaw,
		p.CoatPattern.Source, p.CoatPattern.SystemCoatPatternID, p.CoatPattern.CustomCoatPatternName,
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
			breed_source, system_breed_id, custom_breed_name,
			colors,
			coat_pattern_source, system_coat_pattern_id, custom_coat_pattern_name,
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
	return pet, nil
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
			breed_source, system_breed_id, custom_breed_name,
			colors,
			coat_pattern_source, system_coat_pattern_id, custom_coat_pattern_name,
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

	return items, total, nil
}

func (r *PetRepository) Update(ctx context.Context, petID uuid.UUID, rowVersion int, pet model.Pet) (*model.Pet, error) {
	colorsRaw, err := json.Marshal(pet.Colors)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE pets
		SET name = $3,
		    species_id = $4,
		    sex = $5,
		    birth_date = $6,
		    breed_source = $7,
		    system_breed_id = $8,
		    custom_breed_name = $9,
		    colors = $10,
		    coat_pattern_source = $11,
		    system_coat_pattern_id = $12,
		    custom_coat_pattern_name = $13,
		    is_neutered = $14,
		    is_outdoor = $15,
		    profile_photo_file_id = $16,
		    microchip_id = $17,
		    microchip_installed_at = $18,
		    updated_at = NOW(),
		    row_version = row_version + 1
		WHERE id = $1
		  AND row_version = $2
	`

	cmd, err := r.db.Exec(ctx, query,
		petID, rowVersion,
		pet.Name,
		pet.SpeciesID,
		pet.Sex,
		pet.BirthDate,
		pet.Breed.Source,
		pet.Breed.SystemBreedID,
		pet.Breed.CustomBreedName,
		colorsRaw,
		pet.CoatPattern.Source,
		pet.CoatPattern.SystemCoatPatternID,
		pet.CoatPattern.CustomCoatPatternName,
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

	return r.GetByID(ctx, petID)
}

func (r *PetRepository) UpdatePhoto(ctx context.Context, petID uuid.UUID, rowVersion int, fileID uuid.UUID) (*model.Pet, error) {
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
	var (
		pet       model.Pet
		colorsRaw []byte
	)

	err := s.Scan(
		&pet.ID, &pet.OwnerUserID, &pet.RowVersion, &pet.Name, &pet.SpeciesID, &pet.Sex, &pet.BirthDate,
		&pet.Breed.Source, &pet.Breed.SystemBreedID, &pet.Breed.CustomBreedName,
		&colorsRaw,
		&pet.CoatPattern.Source, &pet.CoatPattern.SystemCoatPatternID, &pet.CoatPattern.CustomCoatPatternName,
		&pet.IsNeutered, &pet.IsOutdoor, &pet.ProfilePhotoFileID,
		&pet.MicrochipID, &pet.MicrochipInstalledAt,
		&pet.Status, &pet.MissingSince, &pet.ArchivedAt,
		&pet.CreatedAt, &pet.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(colorsRaw) == 0 {
		pet.Colors = []model.Color{}
	} else if err := json.Unmarshal(colorsRaw, &pet.Colors); err != nil {
		return nil, err
	}

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

var _ repo.PetRepository = (*PetRepository)(nil)
