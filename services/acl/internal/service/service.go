package service

import (
	"acl/internal/model"
	"acl/internal/repository"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

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
	roles       repository.RoleRepository
	presets     repository.PresetRepository
	invites     repository.InviteRepository
	opts        Options
}

type Options struct {
	InviteTTL          time.Duration
	InviteDeeplinkBase string
}

func New(
	memberships repository.MembershipRepository,
	roles repository.RoleRepository,
	presets repository.PresetRepository,
	invites repository.InviteRepository,
	opts Options,
) *ACLService {
	if opts.InviteTTL <= 0 {
		opts.InviteTTL = 7 * 24 * time.Hour
	}
	if opts.InviteDeeplinkBase == "" {
		opts.InviteDeeplinkBase = "myapp://invite?token="
	}

	return &ACLService{
		memberships: memberships,
		roles:       roles,
		presets:     presets,
		invites:     invites,
		opts:        opts,
	}
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

type CreateInviteParams struct {
	PetID           uuid.UUID
	CreatedByUserID uuid.UUID
	RoleID          uuid.UUID
	Policy          model.Policy
	BasePresetID    *uuid.UUID
}

type CreateInviteResult struct {
	Invite      *repository.InviteView
	DeeplinkURL string
}

type AcceptInviteResult struct {
	PetID  uuid.UUID
	Member *repository.MemberView
}

type CreateCustomRoleParams struct {
	PetID       uuid.UUID
	RequesterID uuid.UUID
	Title       string
}

type UpdateMemberPermissionsParams struct {
	PetID        uuid.UUID
	RequesterID  uuid.UUID
	MemberID     uuid.UUID
	RoleID       uuid.UUID
	Policy       model.Policy
	BasePresetID *uuid.UUID
}

type RemoveMemberParams struct {
	PetID       uuid.UUID
	RequesterID uuid.UUID
	MemberID    uuid.UUID
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

func (s *ACLService) GetMyAccess(ctx context.Context, petID, userID uuid.UUID) (*repository.MemberView, error) {
	member, err := s.memberships.GetActiveViewByPetAndUser(ctx, petID, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	return member, nil
}

func (s *ACLService) ListMembers(ctx context.Context, petID, requesterID uuid.UUID) ([]repository.MemberView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersView),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	return s.memberships.ListActiveViewsByPet(ctx, petID)
}

func (s *ACLService) ListRoles(ctx context.Context, petID, requesterID uuid.UUID) ([]repository.RoleView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersView),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	return s.roles.ListSystemAndPetRoles(ctx, petID)
}

func (s *ACLService) ListPresets(ctx context.Context) ([]repository.PermissionPresetView, error) {
	return s.presets.ListSystem(ctx)
}

func (s *ACLService) CreateCustomRole(ctx context.Context, p CreateCustomRoleParams) (*repository.RoleView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  p.PetID,
		UserID: p.RequesterID,
		Action: string(ActionMembersEditPermissions),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, ErrInvalidInput
	}

	role, err := s.roles.CreateCustom(ctx, p.PetID, title, p.RequesterID)
	if err != nil {
		switch err {
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}
	return role, nil
}

func (s *ACLService) DeleteCustomRole(ctx context.Context, petID, requesterID, roleID uuid.UUID) error {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersEditPermissions),
	})
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	err = s.roles.DeleteCustomIfUnused(ctx, petID, roleID)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return ErrNotFound
		case repository.ErrConflict:
			return ErrConflict
		default:
			return err
		}
	}
	return nil
}

func (s *ACLService) CreateInvite(ctx context.Context, p CreateInviteParams) (*CreateInviteResult, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  p.PetID,
		UserID: p.CreatedByUserID,
		Action: string(ActionMembersInvite),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	role, err := s.roles.GetByID(ctx, p.RoleID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !roleAllowedForPet(role, p.PetID) {
		return nil, ErrInvalidInput
	}

	if p.BasePresetID != nil {
		exists, err := s.presets.ExistsByID(ctx, *p.BasePresetID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidInput
		}
	}

	var (
		invite   *repository.InviteView
		rawToken string
	)
	for i := 0; i < 5; i++ {
		code, err := generateNumericCode(6)
		if err != nil {
			return nil, err
		}
		rawToken, err = generateToken(32)
		if err != nil {
			return nil, err
		}

		invite, err = s.invites.Create(ctx, repository.InviteCreateInput{
			ID:              uuid.New(),
			PetID:           p.PetID,
			CreatedByUserID: p.CreatedByUserID,
			Status:          "ACTIVE",
			TokenHash:       sha256Hex(rawToken),
			Code:            code,
			ExpiresAt:       time.Now().UTC().Add(s.opts.InviteTTL),
			RoleID:          p.RoleID,
			Policy:          p.Policy,
			BasePresetID:    p.BasePresetID,
		})
		if err == nil {
			break
		}
		if err != repository.ErrConflict {
			return nil, err
		}
	}
	if invite == nil {
		return nil, ErrConflict
	}

	return &CreateInviteResult{
		Invite:      invite,
		DeeplinkURL: s.opts.InviteDeeplinkBase + rawToken,
	}, nil
}

func (s *ACLService) ListInvites(ctx context.Context, petID, requesterID uuid.UUID) ([]repository.InviteView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersView),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	return s.invites.ListActiveByPet(ctx, petID)
}

func (s *ACLService) AcceptInviteByCode(ctx context.Context, code string, acceptedByUserID uuid.UUID) (*AcceptInviteResult, error) {
	if len(code) != 6 {
		return nil, ErrInvalidInput
	}
	member, petID, err := s.invites.AcceptByCode(ctx, code, acceptedByUserID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		if err == repository.ErrConflict {
			return nil, ErrConflict
		}
		return nil, err
	}
	return &AcceptInviteResult{PetID: petID, Member: member}, nil
}

func (s *ACLService) AcceptInviteByToken(ctx context.Context, token string, acceptedByUserID uuid.UUID) (*AcceptInviteResult, error) {
	if token == "" {
		return nil, ErrInvalidInput
	}
	member, petID, err := s.invites.AcceptByTokenHash(ctx, sha256Hex(token), acceptedByUserID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		if err == repository.ErrConflict {
			return nil, ErrConflict
		}
		return nil, err
	}
	return &AcceptInviteResult{PetID: petID, Member: member}, nil
}

func (s *ACLService) UpdateMemberPermissions(ctx context.Context, p UpdateMemberPermissionsParams) (*repository.MemberView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  p.PetID,
		UserID: p.RequesterID,
		Action: string(ActionMembersEditPermissions),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	role, err := s.roles.GetByID(ctx, p.RoleID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if !roleAllowedForPet(role, p.PetID) {
		return nil, ErrInvalidInput
	}

	if p.BasePresetID != nil {
		exists, err := s.presets.ExistsByID(ctx, *p.BasePresetID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidInput
		}
	}

	current, err := s.memberships.GetByIDAndPet(ctx, p.PetID, p.MemberID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if current.Status != membershipStatusActive {
		return nil, ErrConflict
	}
	if current.IsPrimaryOwner && !criticalOwnerPolicy(p.Policy) {
		return nil, ErrConflict
	}

	updated, err := s.memberships.UpdatePermissions(ctx, p.PetID, p.MemberID, p.RoleID, p.Policy, p.BasePresetID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (s *ACLService) RemoveMember(ctx context.Context, p RemoveMemberParams) (*repository.MemberView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  p.PetID,
		UserID: p.RequesterID,
		Action: string(ActionMembersRemove),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	member, err := s.memberships.RemoveMember(ctx, p.PetID, p.MemberID, p.RequesterID)
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
	return member, nil
}

func (s *ACLService) RevokeInvite(ctx context.Context, petID, requesterID, inviteID uuid.UUID) error {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersInvite),
	})
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	err = s.invites.RevokeByID(ctx, petID, inviteID)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return ErrNotFound
		case repository.ErrConflict:
			return ErrConflict
		default:
			return err
		}
	}
	return nil
}

func roleAllowedForPet(role *repository.RoleView, petID uuid.UUID) bool {
	switch role.Kind {
	case "SYSTEM":
		return true
	case "CUSTOM":
		return role.PetID != nil && *role.PetID == petID
	default:
		return false
	}
}

func criticalOwnerPolicy(p model.Policy) bool {
	return p.PetRead &&
		p.PetEdit &&
		p.PetStatusChange &&
		p.PetDelete &&
		p.MembersView &&
		p.MembersInvite &&
		p.MembersRemove &&
		p.MembersEditPermissions
}

func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big10)
		if err != nil {
			return "", err
		}
		b[i] = byte('0' + n.Int64())
	}
	return string(b), nil
}

func generateToken(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("invalid token length")
	}
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

var big10 = func() *big.Int {
	return big.NewInt(10)
}()

func sha256Hex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
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
