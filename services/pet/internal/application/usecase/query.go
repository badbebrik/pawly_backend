package usecase

import (
	"context"
	"pet/internal/application/ports"
	"pet/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

func (s *Pet) ListPets(ctx context.Context, p ListPetsParams) ([]PetListItem, int, error) {
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

	memberships, err := s.acl.ListPetsForUser(ctx, p.UserID)
	if err != nil {
		return nil, 0, err
	}
	petIDs := make([]uuid.UUID, 0, len(memberships))
	accessByPet := make(map[uuid.UUID]ACLMembership, len(memberships))
	for i := range memberships {
		member := memberships[i]
		petIDs = append(petIDs, member.PetID)
		if member.MemberID != uuid.Nil {
			accessByPet[member.PetID] = member
		}
	}

	items, total, err := s.repo.ListByIDs(ctx, petIDs, p.IncludeArchived, p.Offset, p.Limit)
	if err != nil {
		return nil, 0, err
	}

	photoIDs := make([]uuid.UUID, 0, len(items))
	for i := range items {
		if items[i].ProfilePhotoFileID != nil {
			photoIDs = append(photoIDs, *items[i].ProfilePhotoFileID)
		}
	}

	photoURLs := make(map[uuid.UUID]string, len(photoIDs))
	if len(photoIDs) > 0 {
		if urls, err := s.file.BatchGetDownloadURLs(ctx, photoIDs); err == nil {
			photoURLs = urls
		}
	}

	out := make([]PetListItem, 0, len(items))
	for i := range items {
		item := PetListItem{
			Pet:      items[i],
			MyAccess: nil,
		}
		if access, ok := accessByPet[items[i].ID]; ok {
			accessCopy := access
			item.MyAccess = &accessCopy
		}
		if items[i].ProfilePhotoFileID != nil {
			if url, ok := photoURLs[*items[i].ProfilePhotoFileID]; ok && strings.TrimSpace(url) != "" {
				urlCopy := url
				item.ProfilePhotoDownloadURL = &urlCopy
			}
		}
		out = append(out, item)
	}

	return out, total, nil
}

func (s *Pet) BatchGetBrief(ctx context.Context, petIDs []uuid.UUID) (map[uuid.UUID]PetBrief, []uuid.UUID, error) {
	if len(petIDs) == 0 {
		return map[uuid.UUID]PetBrief{}, []uuid.UUID{}, nil
	}

	items, _, err := s.repo.ListByIDs(ctx, petIDs, true, 0, len(petIDs))
	if err != nil {
		return nil, nil, err
	}

	photoIDs := make([]uuid.UUID, 0, len(items))
	for i := range items {
		if items[i].ProfilePhotoFileID != nil {
			photoIDs = append(photoIDs, *items[i].ProfilePhotoFileID)
		}
	}

	photoURLs := make(map[uuid.UUID]string, len(photoIDs))
	if len(photoIDs) > 0 {
		if urls, err := s.file.BatchGetDownloadURLs(ctx, photoIDs); err == nil {
			photoURLs = urls
		}
	}

	result := make(map[uuid.UUID]PetBrief, len(items))
	for i := range items {
		var avatarURL *string
		if items[i].ProfilePhotoFileID != nil {
			if url, ok := photoURLs[*items[i].ProfilePhotoFileID]; ok && strings.TrimSpace(url) != "" {
				value := url
				avatarURL = &value
			}
		}

		result[items[i].ID] = PetBrief{
			PetID:     items[i].ID,
			Name:      items[i].Name,
			AvatarURL: avatarURL,
		}
	}

	notFound := make([]uuid.UUID, 0)
	for i := range petIDs {
		if _, ok := result[petIDs[i]]; !ok {
			notFound = append(notFound, petIDs[i])
		}
	}

	return result, notFound, nil
}

func (s *Pet) ResolveProfilePhotoDownloadURL(ctx context.Context, fileID *uuid.UUID) *string {
	if fileID == nil {
		return nil
	}

	url, _, err := s.file.GetDownloadURL(ctx, *fileID)
	if err != nil || strings.TrimSpace(url) == "" {
		return nil
	}

	return &url
}

func (s *Pet) GetPet(ctx context.Context, userID, petID uuid.UUID) (*model.Pet, error) {
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
		if err == ports.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return pet, nil
}

func (s *Pet) GetDictionaries(ctx context.Context) (*model.Dictionaries, error) {
	species, err := s.repo.ListSpecies(ctx)
	if err != nil {
		return nil, err
	}
	breeds, err := s.repo.ListBreeds(ctx)
	if err != nil {
		return nil, err
	}
	patterns, err := s.repo.ListPatterns(ctx)
	if err != nil {
		return nil, err
	}
	colorPresets, err := s.repo.ListColorPresets(ctx)
	if err != nil {
		return nil, err
	}

	return &model.Dictionaries{
		Species:      species,
		Breeds:       breeds,
		Patterns:     patterns,
		ColorPresets: colorPresets,
	}, nil
}
