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
