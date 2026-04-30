package handlers

import (
	"acl/internal/application/ports"
	acluc "acl/internal/application/usecase"
	"acl/internal/transport/http/dto"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func memberToResponse(member *ports.MemberView, profile *ports.ProfileBrief) dto.MemberResponse {
	return dto.MemberResponse{
		ID:             member.ID,
		PetID:          member.PetID,
		UserID:         member.UserID,
		Status:         member.Status,
		IsPrimaryOwner: member.IsPrimaryOwner,
		Role:           roleToResponse(member.Role),
		Policy:         member.Policy,
		Profile:        profileToResponse(profile),
		CreatedAt:      member.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      member.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func profileToResponse(profile *ports.ProfileBrief) *dto.ProfileBriefResponse {
	if profile == nil {
		return nil
	}
	return &dto.ProfileBriefResponse{
		FirstName:         profile.FirstName,
		LastName:          profile.LastName,
		DisplayName:       profile.DisplayName,
		AvatarDownloadURL: profile.AvatarDownloadURL,
	}
}

func roleToResponse(role ports.RoleView) dto.RoleResponse {
	var code *string
	if strings.TrimSpace(role.Code) != "" {
		value := role.Code
		code = &value
	}
	return dto.RoleResponse{
		ID:              role.ID,
		Kind:            role.Kind,
		PetID:           role.PetID,
		Code:            code,
		Title:           role.Title,
		Policy:          role.Policy,
		CreatedByUserID: role.CreatedByUserID,
	}
}

func inviteToResponse(invite ports.InviteView, deeplink string) dto.InviteResponse {
	var deeplinkURL *string
	if strings.TrimSpace(deeplink) != "" {
		deeplinkURL = &deeplink
	}
	var consumedAt *string
	if invite.ConsumedAt != nil {
		value := invite.ConsumedAt.UTC().Format(time.RFC3339)
		consumedAt = &value
	}

	return dto.InviteResponse{
		ID:               invite.ID,
		PetID:            invite.PetID,
		Status:           invite.Status,
		Code:             invite.Code,
		DeeplinkURL:      deeplinkURL,
		ExpiresAt:        invite.ExpiresAt.UTC().Format(time.RFC3339),
		Role:             roleToResponse(invite.Role),
		Policy:           invite.Policy,
		CreatedByUserID:  invite.CreatedByUserID,
		CreatedAt:        invite.CreatedAt.UTC().Format(time.RFC3339),
		ConsumedAt:       consumedAt,
		ConsumedByUserID: invite.ConsumedByUserID,
	}
}

func petBriefToResponse(pet *ports.PetBrief) *dto.PetBriefResponse {
	if pet == nil {
		return nil
	}
	return &dto.PetBriefResponse{
		ID:               pet.PetID,
		Name:             pet.Name,
		PhotoDownloadURL: pet.PhotoDownloadURL,
	}
}

func membersToResponse(items []ports.MemberView, profilesByUser map[uuid.UUID]*ports.ProfileBrief) []dto.MemberResponse {
	out := make([]dto.MemberResponse, 0, len(items))
	for i := range items {
		out = append(out, memberToResponse(&items[i], profilesByUser[items[i].UserID]))
	}
	return out
}

func rolesToResponse(items []ports.RoleView) []dto.RoleResponse {
	out := make([]dto.RoleResponse, 0, len(items))
	for i := range items {
		out = append(out, roleToResponse(items[i]))
	}
	return out
}

func invitesToResponse(items []ports.InviteView) []dto.InviteResponse {
	out := make([]dto.InviteResponse, 0, len(items))
	for i := range items {
		out = append(out, inviteToResponse(items[i], inviteDeeplink("", items[i].Token)))
	}
	return out
}

func invitesToResponseWithBase(items []ports.InviteView, deeplinkBase string) []dto.InviteResponse {
	out := make([]dto.InviteResponse, 0, len(items))
	for i := range items {
		out = append(out, inviteToResponse(items[i], inviteDeeplink(deeplinkBase, items[i].Token)))
	}
	return out
}

func inviteDeeplink(base, token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	base = strings.TrimSpace(base)
	if base == "" {
		base = "pawly://invite?token="
	}
	return base + token
}

func (h *PublicHandlers) loadPetBrief(ctx context.Context, petID uuid.UUID) *ports.PetBrief {
	if h.pets == nil || petID == uuid.Nil {
		return nil
	}

	items, _, err := h.pets.BatchGetBrief(ctx, []uuid.UUID{petID})
	if err != nil || len(items) == 0 {
		return nil
	}

	for i := range items {
		if items[i].PetID == petID {
			item := items[i]
			return &item
		}
	}
	return nil
}

func (h *PublicHandlers) loadProfilesByUserID(ctx context.Context, members []ports.MemberView) map[uuid.UUID]*ports.ProfileBrief {
	result := map[uuid.UUID]*ports.ProfileBrief{}
	if h.profiles == nil || len(members) == 0 {
		return result
	}

	userIDs := make([]uuid.UUID, 0, len(members))
	seen := make(map[uuid.UUID]struct{}, len(members))
	for i := range members {
		userID := members[i].UserID
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}
	if len(userIDs) == 0 {
		return result
	}

	items, _, err := h.profiles.BatchProfilesBrief(ctx, userIDs)
	if err != nil {
		return result
	}
	for i := range items {
		item := items[i]
		itemCopy := item
		result[item.UserID] = &itemCopy
	}
	return result
}

func bootstrapToResponse(petID uuid.UUID, res *acluc.BootstrapResult, profilesByUser map[uuid.UUID]*ports.ProfileBrief, inviteDeeplinkBase string) dto.BootstrapResponse {
	return dto.BootstrapResponse{
		PetID:        petID,
		Me:           memberToResponse(res.Me, profilesByUser[res.Me.UserID]),
		Capabilities: dto.CapabilitiesResponse{MembersRead: res.CanMembersRead, MembersWrite: res.CanMembersWrite},
		Members:      membersToResponse(res.Members, profilesByUser),
		Roles:        rolesToResponse(res.Roles),
		Invites:      invitesToResponseWithBase(res.Invites, inviteDeeplinkBase),
	}
}
