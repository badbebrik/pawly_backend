package handlers

import (
	"pet/internal/application/usecase"
	"pet/internal/domain/model"
	"pet/internal/transport/http/dto"
	"time"

	"github.com/google/uuid"
)

func petToResponse(p *model.Pet, profilePhotoDownloadURL *string) dto.PetResponse {
	return dto.PetResponse{
		ID:                      p.ID,
		OwnerUserID:             p.OwnerUserID,
		RowVersion:              p.RowVersion,
		Name:                    p.Name,
		SpeciesID:               p.SpeciesID,
		SpeciesName:             p.SpeciesName,
		CustomSpeciesName:       p.CustomSpeciesName,
		Sex:                     p.Sex,
		BirthDate:               dateOrNil(p.BirthDate),
		BreedID:                 p.BreedID,
		CustomBreedName:         p.CustomBreedName,
		Colors:                  colorsToResponse(p.Colors),
		PatternID:               p.PatternID,
		CustomPatternName:       p.CustomPatternName,
		IsNeutered:              p.IsNeutered,
		IsOutdoor:               p.IsOutdoor,
		ProfilePhotoFileID:      p.ProfilePhotoFileID,
		ProfilePhotoDownloadURL: profilePhotoDownloadURL,
		MicrochipID:             p.MicrochipID,
		MicrochipInstalledAt:    dateOrNil(p.MicrochipInstalledAt),
		Status:                  p.Status,
		MissingSince:            tsOrNil(p.MissingSince),
		ArchivedAt:              tsOrNil(p.ArchivedAt),
		CreatedAt:               p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:               p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func colorsToResponse(colors []model.Color) []dto.ColorResponse {
	if len(colors) == 0 {
		return []dto.ColorResponse{}
	}

	out := make([]dto.ColorResponse, 0, len(colors))
	for i := range colors {
		id := colors[i].ID
		out = append(out, dto.ColorResponse{
			ID:         &id,
			PresetID:   colors[i].PresetID,
			CustomName: colors[i].CustomName,
			CustomHex:  colors[i].CustomHex,
			SortOrder:  colors[i].SortOrder,
		})
	}

	return out
}

func accessToResponse(access *usecase.ACLMembership) *dto.ACLMembershipResponse {
	if access == nil || access.MemberID == uuid.Nil {
		return nil
	}

	var roleID *uuid.UUID
	if access.Role.ID != uuid.Nil {
		id := access.Role.ID
		roleID = &id
	}

	return &dto.ACLMembershipResponse{
		MemberID:       access.MemberID,
		Status:         access.Status,
		IsPrimaryOwner: access.IsPrimaryOwner,
		Role: dto.ACLRoleResponse{
			ID:              roleID,
			Kind:            access.Role.Kind,
			PetID:           access.Role.PetID,
			Code:            access.Role.Code,
			Title:           access.Role.Title,
			Policy:          aclPolicyToResponse(access.Role.Policy),
			CreatedByUserID: access.Role.CreatedByUserID,
		},
		Policy: aclPolicyToResponse(access.Policy),
	}
}

func aclPolicyToResponse(policy usecase.ACLPolicy) dto.ACLPolicyResponse {
	return dto.ACLPolicyResponse{
		PetRead:      policy.PetRead,
		PetWrite:     policy.PetWrite,
		LogRead:      policy.LogRead,
		LogWrite:     policy.LogWrite,
		HealthRead:   policy.HealthRead,
		HealthWrite:  policy.HealthWrite,
		MembersRead:  policy.MembersRead,
		MembersWrite: policy.MembersWrite,
	}
}

func uploadInfoToResponse(upload usecase.UploadInfo) dto.UploadInfoResponse {
	return dto.UploadInfoResponse{
		Method:    upload.Method,
		URL:       upload.URL,
		Headers:   upload.Headers,
		ExpiresAt: upload.ExpiresAt.UTC().Format(time.RFC3339),
	}
}

func listPetsToResponse(items []usecase.PetListItem, total, offset, limit int) dto.ListPetsResponse {
	out := make([]dto.PetListItemResponse, 0, len(items))
	for i := range items {
		out = append(out, dto.PetListItemResponse{
			Pet:      petToResponse(&items[i].Pet, items[i].ProfilePhotoDownloadURL),
			MyAccess: accessToResponse(items[i].MyAccess),
		})
	}

	return dto.ListPetsResponse{
		Items:  out,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}
}

func dictionariesToResponse(data *model.Dictionaries) dto.DictionariesResponse {
	species := make([]dto.SpeciesResponse, 0, len(data.Species))
	for i := range data.Species {
		species = append(species, dto.SpeciesResponse{
			ID:        data.Species[i].ID,
			Code:      data.Species[i].Code,
			Name:      data.Species[i].Name,
			IconKey:   data.Species[i].IconKey,
			SortOrder: data.Species[i].SortOrder,
			IsActive:  data.Species[i].IsActive,
		})
	}

	breeds := make([]dto.BreedResponse, 0, len(data.Breeds))
	for i := range data.Breeds {
		breeds = append(breeds, dto.BreedResponse{
			ID:        data.Breeds[i].ID,
			SpeciesID: data.Breeds[i].SpeciesID,
			Name:      data.Breeds[i].Name,
			SortOrder: data.Breeds[i].SortOrder,
			IsActive:  data.Breeds[i].IsActive,
		})
	}

	patterns := make([]dto.PatternResponse, 0, len(data.Patterns))
	for i := range data.Patterns {
		patterns = append(patterns, dto.PatternResponse{
			ID:        data.Patterns[i].ID,
			SpeciesID: data.Patterns[i].SpeciesID,
			Name:      data.Patterns[i].Name,
			IconKey:   data.Patterns[i].IconKey,
			SortOrder: data.Patterns[i].SortOrder,
			IsActive:  data.Patterns[i].IsActive,
		})
	}

	colorPresets := make([]dto.ColorPresetResponse, 0, len(data.ColorPresets))
	for i := range data.ColorPresets {
		colorPresets = append(colorPresets, dto.ColorPresetResponse{
			ID:        data.ColorPresets[i].ID,
			Name:      data.ColorPresets[i].Name,
			Hex:       data.ColorPresets[i].Hex,
			SortOrder: data.ColorPresets[i].SortOrder,
			IsActive:  data.ColorPresets[i].IsActive,
		})
	}

	return dto.DictionariesResponse{
		Species:      species,
		Breeds:       breeds,
		Patterns:     patterns,
		ColorPresets: colorPresets,
	}
}

func dateOrNil(v *time.Time) *string {
	if v == nil {
		return nil
	}
	value := v.UTC().Format("2006-01-02")
	return &value
}

func tsOrNil(v *time.Time) *string {
	if v == nil {
		return nil
	}
	value := v.UTC().Format(time.RFC3339)
	return &value
}
