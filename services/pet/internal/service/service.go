package service

import (
	"context"
	"pet/internal/model"
	"pet/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

const ActionPetRead = "pet_read"
const ActionPetEdit = "pet_edit"
const ActionPetStatusChange = "pet_status_change"

type ACLClient interface {
	Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error)
	ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	CreateOwnerMembership(ctx context.Context, petID, userID uuid.UUID) (uuid.UUID, error)
}

type PetService struct {
	repo repository.PetRepository
	acl  ACLClient
}

func New(repo repository.PetRepository, acl ACLClient) *PetService {
	return &PetService{repo: repo, acl: acl}
}

type CreatePetParams struct {
	UserID               uuid.UUID
	Name                 string
	SpeciesID            uuid.UUID
	Sex                  string
	BirthDate            *time.Time
	Breed                model.Breed
	Colors               []model.Color
	CoatPattern          model.CoatPattern
	IsNeutered           string
	IsOutdoor            bool
	ProfilePhotoFileID   *uuid.UUID
	MicrochipID          *string
	MicrochipInstalledAt *time.Time
}

type ListPetsParams struct {
	UserID          uuid.UUID
	IncludeArchived bool
	Offset          int
	Limit           int
}

type UpdatePetParams struct {
	UserID               uuid.UUID
	PetID                uuid.UUID
	RowVersion           int
	Name                 string
	SpeciesID            uuid.UUID
	Sex                  string
	BirthDate            *time.Time
	Breed                model.Breed
	Colors               []model.Color
	CoatPattern          model.CoatPattern
	IsNeutered           string
	IsOutdoor            bool
	ProfilePhotoFileID   *uuid.UUID
	MicrochipID          *string
	MicrochipInstalledAt *time.Time
}

type ChangePetStatusParams struct {
	UserID       uuid.UUID
	PetID        uuid.UUID
	RowVersion   int
	Status       string
	MissingSince *time.Time
}

func (s *PetService) CreatePet(ctx context.Context, p CreatePetParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.SpeciesID == uuid.Nil || strings.TrimSpace(p.Name) == "" {
		return nil, ErrInvalidInput
	}

	pet := model.Pet{
		ID:                   uuid.New(),
		OwnerUserID:          p.UserID,
		Name:                 strings.TrimSpace(p.Name),
		SpeciesID:            p.SpeciesID,
		Sex:                  p.Sex,
		BirthDate:            p.BirthDate,
		Breed:                p.Breed,
		Colors:               p.Colors,
		CoatPattern:          p.CoatPattern,
		IsNeutered:           p.IsNeutered,
		IsOutdoor:            p.IsOutdoor,
		ProfilePhotoFileID:   p.ProfilePhotoFileID,
		MicrochipID:          p.MicrochipID,
		MicrochipInstalledAt: p.MicrochipInstalledAt,
		Status:               "ACTIVE",
	}

	created, err := s.repo.Create(ctx, repository.CreatePetInput{Pet: pet})
	if err != nil {
		if err == repository.ErrConflict {
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

func (s *PetService) ListPets(ctx context.Context, p ListPetsParams) ([]model.Pet, int, error) {
	if p.UserID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Limit <= 0 {
		p.Limit = 50
	}
	if p.Limit > 100 {
		p.Limit = 100
	}

	petIDs, err := s.acl.ListPetsForUser(ctx, p.UserID)
	if err != nil {
		return nil, 0, err
	}

	items, total, err := s.repo.ListByIDs(ctx, petIDs, p.IncludeArchived, p.Offset, p.Limit)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *PetService) GetPet(ctx context.Context, userID, petID uuid.UUID) (*model.Pet, error) {
	if userID == uuid.Nil || petID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, petID, userID, ActionPetRead)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	pet, err := s.repo.GetByID(ctx, petID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return pet, nil
}

func (s *PetService) UpdatePet(ctx context.Context, p UpdatePetParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 || p.SpeciesID == uuid.Nil || strings.TrimSpace(p.Name) == "" {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetEdit)
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
		if err == repository.ErrNotFound {
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
		Sex:                  p.Sex,
		BirthDate:            p.BirthDate,
		Breed:                p.Breed,
		Colors:               p.Colors,
		CoatPattern:          p.CoatPattern,
		IsNeutered:           p.IsNeutered,
		IsOutdoor:            p.IsOutdoor,
		ProfilePhotoFileID:   p.ProfilePhotoFileID,
		MicrochipID:          p.MicrochipID,
		MicrochipInstalledAt: p.MicrochipInstalledAt,
	})
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return nil, ErrNotFound
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}
	return updated, nil
}

func (s *PetService) ChangePetStatus(ctx context.Context, p ChangePetStatusParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetStatusChange)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	status := strings.ToUpper(strings.TrimSpace(p.Status))
	var (
		missingSince *time.Time
		archivedAt   *time.Time
	)
	switch status {
	case "ACTIVE":
	case "MISSING":
		if p.MissingSince == nil {
			return nil, ErrInvalidInput
		}
		missingSince = p.MissingSince
	case "ARCHIVED":
		now := time.Now().UTC()
		archivedAt = &now
	default:
		return nil, ErrInvalidInput
	}

	pet, err := s.repo.UpdateStatus(ctx, p.PetID, p.RowVersion, status, missingSince, archivedAt)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return nil, ErrNotFound
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}
	return pet, nil
}
