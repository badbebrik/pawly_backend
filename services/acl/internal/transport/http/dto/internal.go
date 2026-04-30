package dto

import (
	"acl/internal/domain/model"
	"github.com/google/uuid"
)

type IsMemberRequest struct {
	PetID  string `json:"pet_id"`
	UserID string `json:"user_id"`
}

type GetPolicyRequest struct {
	PetID  string `json:"pet_id"`
	UserID string `json:"user_id"`
}

type CheckRequest struct {
	PetID  string `json:"pet_id"`
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

type ListPetsForUserRequest struct {
	UserID string `json:"user_id"`
}

type ListMembersForPetRequest struct {
	PetID string `json:"pet_id"`
}

type TransferOwnershipRequest struct {
	PetID           string `json:"pet_id"`
	RequesterUserID string `json:"requester_user_id"`
	TargetMemberID  string `json:"target_member_id"`
}

type IsMemberResponse struct {
	IsMember bool `json:"is_member"`
}

type GetPolicyResponse struct {
	MemberID       uuid.UUID    `json:"member_id"`
	Status         string       `json:"status"`
	IsPrimaryOwner bool         `json:"is_primary_owner"`
	Policy         model.Policy `json:"policy"`
}

type CheckResponse struct {
	Allowed bool `json:"allowed"`
}

type ListPetsForUserResponse struct {
	PetIDs      []string         `json:"pet_ids"`
	Memberships []MemberResponse `json:"memberships"`
}

type ListMembersForPetResponse struct {
	UserIDs []string         `json:"user_ids"`
	Members []MemberResponse `json:"members"`
}

type TransferOwnershipResponse struct {
	PreviousOwnerMemberID uuid.UUID `json:"previous_owner_member_id"`
	PreviousOwnerUserID   uuid.UUID `json:"previous_owner_user_id"`
	CurrentOwnerMemberID  uuid.UUID `json:"current_owner_member_id"`
	CurrentOwnerUserID    uuid.UUID `json:"current_owner_user_id"`
}
