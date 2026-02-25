package service

import (
	"acl/internal/model"
	"acl/internal/repository"
	"context"
	"strings"

	"github.com/google/uuid"
)

type Action string

const (
	ActionPetRead                Action = "pet_read"
	ActionPetEdit                Action = "pet_edit"
	ActionPetStatusChange        Action = "pet_status_change"
	ActionPetDelete              Action = "pet_delete"
	ActionLogRead                Action = "log_read"
	ActionLogCreate              Action = "log_create"
	ActionLogEdit                Action = "log_edit"
	ActionLogDelete              Action = "log_delete"
	ActionLogAttachmentsRead     Action = "log_attachments_read"
	ActionHealthRead             Action = "health_read"
	ActionHealthWrite            Action = "health_write"
	ActionTaskRead               Action = "task_read"
	ActionTaskWrite              Action = "task_write"
	ActionMembersView            Action = "members_view"
	ActionMembersInvite          Action = "members_invite"
	ActionMembersRemove          Action = "members_remove"
	ActionMembersEditPermissions Action = "members_edit_permissions"
)

const membershipStatusActive = "ACTIVE"

type ACLService struct {
	memberships repository.MembershipRepository
}

func New(memberships repository.MembershipRepository) *ACLService {
	return &ACLService{memberships: memberships}
}

type CheckParams struct {
	PetID  uuid.UUID
	UserID uuid.UUID
	Action string
}

type PolicyResult struct {
	MemberID       uuid.UUID
	Status         string
	IsPrimaryOwner bool
	Policy         model.Policy
}

func (s *ACLService) IsMember(ctx context.Context, petID, userID uuid.UUID) (bool, error) {
	_, err := s.memberships.GetActiveByPetAndUser(ctx, petID, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *ACLService) GetPolicy(ctx context.Context, petID, userID uuid.UUID) (*PolicyResult, error) {
	access, err := s.memberships.GetByPetAndUser(ctx, petID, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if access.Status != membershipStatusActive {
		return nil, ErrForbidden
	}

	return &PolicyResult{
		MemberID:       access.MemberID,
		Status:         access.Status,
		IsPrimaryOwner: access.IsPrimaryOwner,
		Policy:         access.Policy,
	}, nil
}

func (s *ACLService) Check(ctx context.Context, p CheckParams) (bool, error) {
	if p.Action == "" {
		return false, ErrInvalidInput
	}

	policyRes, err := s.GetPolicy(ctx, p.PetID, p.UserID)
	if err != nil {
		return false, err
	}

	allowed, ok := isAllowed(policyRes.Policy, Action(strings.ToLower(p.Action)))
	if !ok {
		return false, ErrInvalidInput
	}
	return allowed, nil
}

func (s *ACLService) ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.memberships.ListActivePetIDsByUser(ctx, userID)
}

func isAllowed(p model.Policy, action Action) (bool, bool) {
	switch action {
	case ActionPetRead:
		return p.PetRead, true
	case ActionPetEdit:
		return p.PetEdit, true
	case ActionPetStatusChange:
		return p.PetStatusChange, true
	case ActionPetDelete:
		return p.PetDelete, true
	case ActionLogRead:
		return p.LogRead, true
	case ActionLogCreate:
		return p.LogCreate, true
	case ActionLogEdit:
		return p.LogEdit, true
	case ActionLogDelete:
		return p.LogDelete, true
	case ActionLogAttachmentsRead:
		return p.LogAttachmentsRead, true
	case ActionHealthRead:
		return p.HealthRead, true
	case ActionHealthWrite:
		return p.HealthWrite, true
	case ActionTaskRead:
		return p.TaskRead, true
	case ActionTaskWrite:
		return p.TaskWrite, true
	case ActionMembersView:
		return p.MembersView, true
	case ActionMembersInvite:
		return p.MembersInvite, true
	case ActionMembersRemove:
		return p.MembersRemove, true
	case ActionMembersEditPermissions:
		return p.MembersEditPermissions, true
	default:
		return false, false
	}
}
