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
	ActionPetRead      Action = "pet_read"
	ActionPetWrite     Action = "pet_write"
	ActionLogRead      Action = "log_read"
	ActionLogWrite     Action = "log_write"
	ActionHealthRead   Action = "health_read"
	ActionHealthWrite  Action = "health_write"
	ActionTaskRead     Action = "task_read"
	ActionTaskWrite    Action = "task_write"
	ActionMembersRead  Action = "members_read"
	ActionMembersWrite Action = "members_write"
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

type RegenerateInviteLinkResult struct {
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

type TransferOwnershipParams struct {
	PetID          uuid.UUID
	RequesterID    uuid.UUID
	TargetMemberID uuid.UUID
}

type TransferOwnershipResult struct {
	PreviousOwner *repository.MemberView
	CurrentOwner  *repository.MemberView
}

type BootstrapResult struct {
	Me              *repository.MemberView
	Members         []repository.MemberView
	Roles           []repository.RoleView
	Presets         []repository.PermissionPresetView
	Invites         []repository.InviteView
	CanMembersRead  bool
	CanMembersWrite bool
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

	allowed, ok := isAllowed(policyRes.Policy, normalizeAction(strings.ToLower(p.Action)))
	if !ok {
		return false, ErrInvalidInput
	}
	return allowed, nil
}

func (s *ACLService) ListPetMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]repository.MemberView, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return s.memberships.ListActiveViewsByUser(ctx, userID)
}

func (s *ACLService) CreateOwnerMembership(ctx context.Context, petID, ownerUserID uuid.UUID) (*repository.MemberView, error) {
	if petID == uuid.Nil || ownerUserID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	member, err := s.memberships.CreateOwner(ctx, petID, ownerUserID, ownerFullAccessPolicy())
	if err != nil {
		switch err {
		case repository.ErrConflict:
			return nil, ErrConflict
		case repository.ErrNotFound:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return member, nil
}

func (s *ACLService) TransferOwnership(ctx context.Context, p TransferOwnershipParams) (*TransferOwnershipResult, error) {
	if p.PetID == uuid.Nil || p.RequesterID == uuid.Nil || p.TargetMemberID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	res, err := s.memberships.TransferOwnership(ctx, p.PetID, p.RequesterID, p.TargetMemberID)
	if err != nil {
		switch err {
		case repository.ErrForbidden:
			return nil, ErrForbidden
		case repository.ErrNotFound:
			return nil, ErrNotFound
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}

	return &TransferOwnershipResult{
		PreviousOwner: &res.PreviousOwner,
		CurrentOwner:  &res.CurrentOwner,
	}, nil
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

func (s *ACLService) LeavePet(ctx context.Context, petID, userID uuid.UUID) (*repository.MemberView, error) {
	if petID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	member, err := s.memberships.GetActiveViewByPetAndUser(ctx, petID, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if member.IsPrimaryOwner {
		return nil, ErrConflict
	}

	removed, err := s.memberships.RemoveMember(ctx, petID, member.ID, userID)
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
	return removed, nil
}

func (s *ACLService) GetBootstrap(ctx context.Context, petID, userID uuid.UUID) (*BootstrapResult, error) {
	if petID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	me, err := s.GetMyAccess(ctx, petID, userID)
	if err != nil {
		return nil, err
	}
	if !me.Policy.MembersRead {
		return nil, ErrForbidden
	}

	members, err := s.memberships.ListActiveViewsByPet(ctx, petID)
	if err != nil {
		return nil, err
	}
	roles, err := s.roles.ListSystemAndPetRoles(ctx, petID)
	if err != nil {
		return nil, err
	}
	presets, err := s.presets.ListSystem(ctx)
	if err != nil {
		return nil, err
	}
	invites, err := s.invites.ListActiveByPet(ctx, petID)
	if err != nil {
		return nil, err
	}

	return &BootstrapResult{
		Me:              me,
		Members:         members,
		Roles:           roles,
		Presets:         presets,
		Invites:         invites,
		CanMembersRead:  me.Policy.MembersRead,
		CanMembersWrite: me.Policy.MembersWrite,
	}, nil
}

func (s *ACLService) ListMembers(ctx context.Context, petID, requesterID uuid.UUID) ([]repository.MemberView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersRead),
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
		Action: string(ActionMembersRead),
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
		Action: string(ActionMembersWrite),
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
		Action: string(ActionMembersWrite),
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
		Action: string(ActionMembersWrite),
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
			TokenValue:      rawToken,
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

func (s *ACLService) RegenerateInviteLink(ctx context.Context, petID, requesterID, inviteID uuid.UUID) (*RegenerateInviteLinkResult, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersWrite),
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	var (
		invite   *repository.InviteView
		rawToken string
	)
	for i := 0; i < 5; i++ {
		rawToken, err = generateToken(32)
		if err != nil {
			return nil, err
		}
		invite, err = s.invites.RotateTokenHashByID(ctx, petID, inviteID, sha256Hex(rawToken), rawToken)
		if err == nil {
			break
		}
		if err != repository.ErrConflict {
			if err == repository.ErrNotFound {
				return nil, ErrNotFound
			}
			return nil, err
		}
	}
	if invite == nil {
		return nil, ErrConflict
	}

	return &RegenerateInviteLinkResult{
		Invite:      invite,
		DeeplinkURL: s.opts.InviteDeeplinkBase + rawToken,
	}, nil
}

func (s *ACLService) ListInvites(ctx context.Context, petID, requesterID uuid.UUID) ([]repository.InviteView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  petID,
		UserID: requesterID,
		Action: string(ActionMembersRead),
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

func (s *ACLService) PreviewInviteByToken(ctx context.Context, token string) (*repository.InviteView, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidInput
	}

	invite, err := s.invites.GetActiveByTokenHash(ctx, sha256Hex(token))
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return invite, nil
}

func (s *ACLService) UpdateMemberPermissions(ctx context.Context, p UpdateMemberPermissionsParams) (*repository.MemberView, error) {
	allowed, err := s.Check(ctx, CheckParams{
		PetID:  p.PetID,
		UserID: p.RequesterID,
		Action: string(ActionMembersWrite),
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
	if current.UserID == p.RequesterID {
		return nil, ErrForbidden
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
		Action: string(ActionMembersWrite),
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
		Action: string(ActionMembersWrite),
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
		p.PetWrite &&
		p.LogRead &&
		p.LogWrite &&
		p.HealthRead &&
		p.HealthWrite &&
		p.TaskRead &&
		p.TaskWrite &&
		p.MembersRead &&
		p.MembersWrite
}

func ownerFullAccessPolicy() model.Policy {
	return model.Policy{
		PetRead:      true,
		PetWrite:     true,
		LogRead:      true,
		LogWrite:     true,
		HealthRead:   true,
		HealthWrite:  true,
		TaskRead:     true,
		TaskWrite:    true,
		MembersRead:  true,
		MembersWrite: true,
	}
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

func normalizeAction(raw string) Action {
	switch Action(raw) {
	case ActionPetRead:
		return ActionPetRead
	case ActionPetWrite:
		return ActionPetWrite
	case ActionLogRead:
		return ActionLogRead
	case ActionLogWrite:
		return ActionLogWrite
	case ActionHealthRead:
		return ActionHealthRead
	case ActionHealthWrite:
		return ActionHealthWrite
	case ActionTaskRead:
		return ActionTaskRead
	case ActionTaskWrite:
		return ActionTaskWrite
	case ActionMembersRead:
		return ActionMembersRead
	case ActionMembersWrite:
		return ActionMembersWrite
	default:
		return ""
	}
}

func isAllowed(p model.Policy, action Action) (bool, bool) {
	switch action {
	case ActionPetRead:
		return p.PetRead, true
	case ActionPetWrite:
		return p.PetWrite, true
	case ActionLogRead:
		return p.LogRead, true
	case ActionLogWrite:
		return p.LogWrite, true
	case ActionHealthRead:
		return p.HealthRead, true
	case ActionHealthWrite:
		return p.HealthWrite, true
	case ActionTaskRead:
		return p.TaskRead, true
	case ActionTaskWrite:
		return p.TaskWrite, true
	case ActionMembersRead:
		return p.MembersRead, true
	case ActionMembersWrite:
		return p.MembersWrite, true
	default:
		return false, false
	}
}
