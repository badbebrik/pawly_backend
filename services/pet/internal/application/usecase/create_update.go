package usecase

import (
	"context"
	"pet/internal/application/ports"
	"pet/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

func (s *Pet) CreatePet(ctx context.Context, p CreatePetParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || strings.TrimSpace(p.Name) == "" {
		return nil, ErrInvalidInput
	}
	colors, err := normalizeColors(p.Colors)
	if err != nil {
		return nil, err
	}
	customSpeciesName, err := normalizeSpeciesChoice(p.SpeciesID, p.CustomSpeciesName)
	if err != nil {
		return nil, err
	}
	if customSpeciesName != nil && p.BreedID != nil {
		return nil, ErrInvalidInput
	}
	customBreedName, err := normalizeExclusiveTextChoice(p.BreedID, p.CustomBreedName)
	if err != nil {
		return nil, err
	}
	customPatternName, err := normalizeExclusiveTextChoice(p.PatternID, p.CustomPatternName)
	if err != nil {
		return nil, err
	}

	pet := model.Pet{
		ID:                   uuid.New(),
		OwnerUserID:          p.UserID,
		Name:                 strings.TrimSpace(p.Name),
		SpeciesID:            p.SpeciesID,
		CustomSpeciesName:    customSpeciesName,
		Sex:                  p.Sex,
		BirthDate:            p.BirthDate,
		BreedID:              p.BreedID,
		CustomBreedName:      customBreedName,
		PatternID:            p.PatternID,
		CustomPatternName:    customPatternName,
		Colors:               colors,
		IsNeutered:           p.IsNeutered,
		IsOutdoor:            p.IsOutdoor,
		MicrochipID:          p.MicrochipID,
		MicrochipInstalledAt: p.MicrochipInstalledAt,
		Status:               "ACTIVE",
	}

	created, err := s.repo.Create(ctx, ports.CreatePetInput{Pet: pet})
	if err != nil {
		if err == ports.ErrConflict {
			return nil, ErrConflict
		}
		return nil, err
	}

	if _, err := s.acl.CreateOwnerMembership(ctx, created.ID, p.UserID); err != nil {
		_ = s.repo.DeleteByID(ctx, created.ID)
		if err == ErrConflict {
			return nil, ErrConflict
		}
		return nil, err
	}

	return created, nil
}

func (s *Pet) UpdatePet(ctx context.Context, p UpdatePetParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 || strings.TrimSpace(p.Name) == "" {
		return nil, ErrInvalidInput
	}
	colors, err := normalizeColors(p.Colors)
	if err != nil {
		return nil, err
	}
	customSpeciesName, err := normalizeSpeciesChoice(p.SpeciesID, p.CustomSpeciesName)
	if err != nil {
		return nil, err
	}
	if customSpeciesName != nil && p.BreedID != nil {
		return nil, ErrInvalidInput
	}
	customBreedName, err := normalizeExclusiveTextChoice(p.BreedID, p.CustomBreedName)
	if err != nil {
		return nil, err
	}
	customPatternName, err := normalizeExclusiveTextChoice(p.PatternID, p.CustomPatternName)
	if err != nil {
		return nil, err
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetWrite)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	current, err := s.repo.GetByID(ctx, p.PetID)
	if err != nil {
		if err == ports.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if current.Status == "ARCHIVED" {
		return nil, ErrConflict
	}

	updated, err := s.repo.Update(ctx, p.PetID, p.RowVersion, model.Pet{
		Name:                 strings.TrimSpace(p.Name),
		SpeciesID:            p.SpeciesID,
		CustomSpeciesName:    customSpeciesName,
		Sex:                  p.Sex,
		BirthDate:            p.BirthDate,
		BreedID:              p.BreedID,
		CustomBreedName:      customBreedName,
		PatternID:            p.PatternID,
		CustomPatternName:    customPatternName,
		Colors:               colors,
		IsNeutered:           p.IsNeutered,
		IsOutdoor:            p.IsOutdoor,
		ProfilePhotoFileID:   current.ProfilePhotoFileID,
		MicrochipID:          p.MicrochipID,
		MicrochipInstalledAt: p.MicrochipInstalledAt,
	})
	if err != nil {
		switch err {
		case ports.ErrNotFound:
			return nil, ErrNotFound
		case ports.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}

	return updated, nil
}
