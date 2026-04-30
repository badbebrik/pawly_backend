package dto

import (
	"acl/internal/domain/model"
	"github.com/google/uuid"
)

type CreateInviteRequest struct {
	RoleID string       `json:"role_id"`
	Policy model.Policy `json:"policy"`
}

type CreateCustomRoleRequest struct {
	Title  string       `json:"title"`
	Policy model.Policy `json:"policy"`
}

type UpdateCustomRoleRequest struct {
	Title  string       `json:"title"`
	Policy model.Policy `json:"policy"`
}

type UpdateMemberPermissionsRequest struct {
	RoleID string       `json:"role_id"`
	Policy model.Policy `json:"policy"`
}

type AcceptInviteByCodeRequest struct {
	Code string `json:"code"`
}

type AcceptInviteByTokenRequest struct {
	Token string `json:"token"`
}

type PreviewInviteByTokenRequest struct {
	Token string `json:"token"`
}

type ProfileBriefResponse struct {
	FirstName         *string `json:"first_name"`
	LastName          *string `json:"last_name"`
	DisplayName       *string `json:"display_name"`
	AvatarDownloadURL *string `json:"avatar_download_url"`
}

type RoleResponse struct {
	ID              uuid.UUID    `json:"id"`
	Kind            string       `json:"kind"`
	PetID           *uuid.UUID   `json:"pet_id"`
	Code            *string      `json:"code"`
	Title           string       `json:"title"`
	Policy          model.Policy `json:"policy"`
	CreatedByUserID *uuid.UUID   `json:"created_by_user_id"`
}

type MemberResponse struct {
	ID             uuid.UUID             `json:"id"`
	PetID          uuid.UUID             `json:"pet_id"`
	UserID         uuid.UUID             `json:"user_id"`
	Status         string                `json:"status"`
	IsPrimaryOwner bool                  `json:"is_primary_owner"`
	Role           RoleResponse          `json:"role"`
	Policy         model.Policy          `json:"policy"`
	Profile        *ProfileBriefResponse `json:"profile"`
	CreatedAt      string                `json:"created_at"`
	UpdatedAt      string                `json:"updated_at"`
}

type InviteResponse struct {
	ID               uuid.UUID    `json:"id"`
	PetID            uuid.UUID    `json:"pet_id"`
	Status           string       `json:"status"`
	Code             string       `json:"code"`
	DeeplinkURL      *string      `json:"deeplink_url,omitempty"`
	ExpiresAt        string       `json:"expires_at"`
	Role             RoleResponse `json:"role"`
	Policy           model.Policy `json:"policy"`
	CreatedByUserID  uuid.UUID    `json:"created_by_user_id"`
	CreatedAt        string       `json:"created_at"`
	ConsumedAt       *string      `json:"consumed_at"`
	ConsumedByUserID *uuid.UUID   `json:"consumed_by_user_id"`
}

type PetBriefResponse struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	PhotoDownloadURL *string   `json:"photo_download_url"`
}

type CapabilitiesResponse struct {
	MembersRead  bool `json:"members_read"`
	MembersWrite bool `json:"members_write"`
}

type GetMyAccessResponse struct {
	PetID  uuid.UUID      `json:"pet_id"`
	Member MemberResponse `json:"member"`
}

type MemberEnvelopeResponse struct {
	Member MemberResponse `json:"member"`
}

type MembersListResponse struct {
	Items []MemberResponse `json:"items"`
}

type RolesListResponse struct {
	Items []RoleResponse `json:"items"`
}

type RoleEnvelopeResponse struct {
	Role RoleResponse `json:"role"`
}

type InviteEnvelopeResponse struct {
	Invite InviteResponse `json:"invite"`
}

type InvitesListResponse struct {
	Items []InviteResponse `json:"items"`
}

type AcceptInviteResponse struct {
	PetID  uuid.UUID      `json:"pet_id"`
	Member MemberResponse `json:"member"`
}

type PreviewInviteResponse struct {
	Invite InviteResponse    `json:"invite"`
	Pet    *PetBriefResponse `json:"pet,omitempty"`
}

type BootstrapResponse struct {
	PetID        uuid.UUID            `json:"pet_id"`
	Me           MemberResponse       `json:"me"`
	Capabilities CapabilitiesResponse `json:"capabilities"`
	Members      []MemberResponse     `json:"members"`
	Roles        []RoleResponse       `json:"roles"`
	Invites      []InviteResponse     `json:"invites"`
}
