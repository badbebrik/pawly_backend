package usecase

import (
	"acl/internal/application/ports"
	"acl/internal/domain/model"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeMembershipRepo struct {
	byPetUser       *ports.MembershipAccess
	activeByPetUser *ports.MembershipAccess
	activeView      *ports.MemberView
	activeViews     []ports.MemberView
	ownerResult     *ports.MemberView
	ownerErr        error
	transferResult  *ports.TransferOwnershipView
	updateResult    *ports.MemberView
	removeResult    *ports.MemberView
	transferErr     error
	updateErr       error
	removeErr       error
	err             error
}

func (f *fakeMembershipRepo) GetByPetAndUser(_ context.Context, _, _ uuid.UUID) (*ports.MembershipAccess, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.byPetUser == nil {
		return nil, ports.ErrNotFound
	}
	return f.byPetUser, nil
}

func (f *fakeMembershipRepo) GetActiveByPetAndUser(_ context.Context, _, _ uuid.UUID) (*ports.MembershipAccess, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.activeByPetUser == nil {
		return nil, ports.ErrNotFound
	}
	return f.activeByPetUser, nil
}

func (f *fakeMembershipRepo) CreateOwner(_ context.Context, _, _ uuid.UUID, _ model.Policy) (*ports.MemberView, error) {
	if f.ownerErr != nil {
		return nil, f.ownerErr
	}
	if f.ownerResult == nil {
		return nil, ports.ErrNotFound
	}
	return f.ownerResult, nil
}

func (f *fakeMembershipRepo) GetActiveViewByPetAndUser(_ context.Context, _, _ uuid.UUID) (*ports.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.activeView == nil {
		return nil, ports.ErrNotFound
	}
	return f.activeView, nil
}

func (f *fakeMembershipRepo) GetByIDAndPet(_ context.Context, _, _ uuid.UUID) (*ports.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.activeView == nil {
		return nil, ports.ErrNotFound
	}
	return f.activeView, nil
}

func (f *fakeMembershipRepo) ListActiveViewsByPet(_ context.Context, _ uuid.UUID) ([]ports.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.activeViews, nil
}

func (f *fakeMembershipRepo) ListActiveViewsByUser(_ context.Context, _ uuid.UUID) ([]ports.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.activeViews, nil
}

func (f *fakeMembershipRepo) TransferOwnership(_ context.Context, _, _, _ uuid.UUID) (*ports.TransferOwnershipView, error) {
	if f.transferErr != nil {
		return nil, f.transferErr
	}
	if f.transferResult == nil {
		return nil, ports.ErrNotFound
	}
	return f.transferResult, nil
}

func (f *fakeMembershipRepo) UpdatePermissions(_ context.Context, _, _ uuid.UUID, _ uuid.UUID, _ model.Policy) (*ports.MemberView, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updateResult == nil {
		return nil, ports.ErrNotFound
	}
	return f.updateResult, nil
}

func (f *fakeMembershipRepo) RemoveMember(_ context.Context, _, _ uuid.UUID, _ uuid.UUID) (*ports.MemberView, error) {
	if f.removeErr != nil {
		return nil, f.removeErr
	}
	if f.removeResult == nil {
		return nil, ports.ErrNotFound
	}
	return f.removeResult, nil
}

type fakeRoleRepo struct {
	roles      []ports.RoleView
	byID       *ports.RoleView
	createRole *ports.RoleView
	updateRole *ports.RoleView
	createErr  error
	updateErr  error
	deleteErr  error
	err        error
}

func (f *fakeRoleRepo) GetByID(_ context.Context, _ uuid.UUID) (*ports.RoleView, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.byID == nil {
		return nil, ports.ErrNotFound
	}
	return f.byID, nil
}

func (f *fakeRoleRepo) ListSystemAndPetRoles(_ context.Context, _ uuid.UUID) ([]ports.RoleView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.roles, nil
}

func (f *fakeRoleRepo) CreateCustom(_ context.Context, _ uuid.UUID, _ string, _ model.Policy, _ uuid.UUID) (*ports.RoleView, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createRole == nil {
		return nil, ports.ErrNotFound
	}
	return f.createRole, nil
}

func (f *fakeRoleRepo) UpdateCustom(_ context.Context, _, _ uuid.UUID, _ string, _ model.Policy) (*ports.RoleView, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updateRole == nil {
		return nil, ports.ErrNotFound
	}
	return f.updateRole, nil
}

func (f *fakeRoleRepo) DeleteCustomIfUnused(_ context.Context, _, _ uuid.UUID) error {
	return f.deleteErr
}

type fakeInviteRepo struct {
	createErrSeq   []error
	createCalls    int
	createInvite   *ports.InviteView
	rotateErrSeq   []error
	rotateCalls    int
	rotateInvite   *ports.InviteView
	previewInvite  *ports.InviteView
	previewErr     error
	revokeErr      error
	acceptCodeErr  error
	acceptTokenErr error
}

func (f *fakeInviteRepo) Create(_ context.Context, _ ports.InviteCreateInput) (*ports.InviteView, error) {
	call := f.createCalls
	f.createCalls++
	if call < len(f.createErrSeq) && f.createErrSeq[call] != nil {
		return nil, f.createErrSeq[call]
	}
	if f.createInvite == nil {
		return &ports.InviteView{ID: uuid.New(), PetID: uuid.New(), Status: "ACTIVE"}, nil
	}
	return f.createInvite, nil
}

func (f *fakeInviteRepo) ListActiveByPet(_ context.Context, _ uuid.UUID) ([]ports.InviteView, error) {
	return []ports.InviteView{}, nil
}

func (f *fakeInviteRepo) GetActiveByToken(_ context.Context, _ string) (*ports.InviteView, error) {
	if f.previewErr != nil {
		return nil, f.previewErr
	}
	if f.previewInvite == nil {
		return nil, ports.ErrNotFound
	}
	return f.previewInvite, nil
}

func (f *fakeInviteRepo) AcceptByCode(_ context.Context, _ string, _ uuid.UUID) (*ports.MemberView, uuid.UUID, error) {
	if f.acceptCodeErr != nil {
		return nil, uuid.Nil, f.acceptCodeErr
	}
	member := &ports.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE"}
	return member, member.PetID, nil
}

func (f *fakeInviteRepo) AcceptByToken(_ context.Context, _ string, _ uuid.UUID) (*ports.MemberView, uuid.UUID, error) {
	if f.acceptTokenErr != nil {
		return nil, uuid.Nil, f.acceptTokenErr
	}
	member := &ports.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE"}
	return member, member.PetID, nil
}

func (f *fakeInviteRepo) RotateTokenByID(_ context.Context, _, _ uuid.UUID, _ string) (*ports.InviteView, error) {
	call := f.rotateCalls
	f.rotateCalls++
	if call < len(f.rotateErrSeq) && f.rotateErrSeq[call] != nil {
		return nil, f.rotateErrSeq[call]
	}
	if f.rotateInvite == nil {
		return &ports.InviteView{ID: uuid.New(), PetID: uuid.New(), Status: "ACTIVE"}, nil
	}
	return f.rotateInvite, nil
}

func (f *fakeInviteRepo) RevokeByID(_ context.Context, _, _ uuid.UUID) error {
	return f.revokeErr
}

func newTestService(m *fakeMembershipRepo, r *fakeRoleRepo, i *fakeInviteRepo) *ACL {
	if i == nil {
		i = &fakeInviteRepo{}
	}
	return newACL(m, r, i, Options{InviteDeeplinkBase: "pawly://invite?token="})
}

func TestCheckAllowed(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		MemberID:       uuid.New(),
		Status:         "ACTIVE",
		IsPrimaryOwner: false,
		Policy:         model.Policy{PetRead: true},
	}}, &fakeRoleRepo{}, nil)

	allowed, err := svc.Check(context.Background(), CheckParams{PetID: uuid.New(), UserID: uuid.New(), Action: string(ActionPetRead)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected allowed=true")
	}
}

func TestCheckInvalidAction(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{Status: "ACTIVE"}}, &fakeRoleRepo{}, nil)

	_, err := svc.Check(context.Background(), CheckParams{PetID: uuid.New(), UserID: uuid.New(), Action: "unknown_action"})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestBasicInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "list pet memberships missing user",
			run: func() error {
				_, err := svc.ListPetMembershipsForUser(context.Background(), uuid.Nil)
				return err
			},
		},
		{
			name: "list active members missing pet",
			run: func() error {
				_, err := svc.ListActiveMembersForPet(context.Background(), uuid.Nil)
				return err
			},
		},
		{
			name: "create owner missing pet",
			run: func() error {
				_, err := svc.CreateOwnerMembership(context.Background(), uuid.Nil, uuid.New())
				return err
			},
		},
		{
			name: "transfer missing target member",
			run: func() error {
				_, err := svc.TransferOwnership(context.Background(), TransferOwnershipParams{
					PetID:       uuid.New(),
					RequesterID: uuid.New(),
				})
				return err
			},
		},
		{
			name: "bootstrap missing user",
			run: func() error {
				_, err := svc.GetBootstrap(context.Background(), uuid.New(), uuid.Nil)
				return err
			},
		},
		{
			name: "leave missing pet",
			run: func() error {
				_, err := svc.LeavePet(context.Background(), uuid.Nil, uuid.New())
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err != ErrInvalidInput {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestIsMember(t *testing.T) {
	t.Parallel()

	memberSvc := newTestService(&fakeMembershipRepo{
		activeByPetUser: &ports.MembershipAccess{Status: "ACTIVE"},
	}, &fakeRoleRepo{}, nil)

	ok, err := memberSvc.IsMember(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected member")
	}

	nonMemberSvc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)
	ok, err = nonMemberSvc.IsMember(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected non-member")
	}
}

func TestGetPolicyReturnsActivePolicy(t *testing.T) {
	t.Parallel()

	memberID := uuid.New()
	policy := model.Policy{PetRead: true, MembersRead: true}
	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{
			MemberID:       memberID,
			Status:         "ACTIVE",
			IsPrimaryOwner: true,
			Policy:         policy,
		},
	}, &fakeRoleRepo{}, nil)

	out, err := svc.GetPolicy(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.MemberID != memberID || !out.IsPrimaryOwner || !out.Policy.PetRead || !out.Policy.MembersRead {
		t.Fatalf("unexpected policy result: %+v", out)
	}
}

func TestGetPolicyRejectsInactiveMembership(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{Status: "REMOVED"},
	}, &fakeRoleRepo{}, nil)

	_, err := svc.GetPolicy(context.Background(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateOwnerMembershipMapsConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		ownerErr: ports.ErrConflict,
	}, &fakeRoleRepo{}, nil)

	_, err := svc.CreateOwnerMembership(context.Background(), uuid.New(), uuid.New())
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestCreateOwnerMembershipReturnsMember(t *testing.T) {
	t.Parallel()

	member := &ports.MemberView{
		ID:             uuid.New(),
		PetID:          uuid.New(),
		UserID:         uuid.New(),
		Status:         "ACTIVE",
		IsPrimaryOwner: true,
		Policy:         ownerFullAccessPolicy(),
	}
	svc := newTestService(&fakeMembershipRepo{ownerResult: member}, &fakeRoleRepo{}, nil)

	out, err := svc.CreateOwnerMembership(context.Background(), member.PetID, member.UserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != member {
		t.Fatalf("unexpected owner member: %+v", out)
	}
}

func TestTransferOwnershipReturnsPreviousAndCurrentOwners(t *testing.T) {
	t.Parallel()

	previous := ports.MemberView{ID: uuid.New(), UserID: uuid.New(), IsPrimaryOwner: false}
	current := ports.MemberView{ID: uuid.New(), UserID: uuid.New(), IsPrimaryOwner: true}
	svc := newTestService(&fakeMembershipRepo{
		transferResult: &ports.TransferOwnershipView{
			PreviousOwner: previous,
			CurrentOwner:  current,
		},
	}, &fakeRoleRepo{}, nil)

	out, err := svc.TransferOwnership(context.Background(), TransferOwnershipParams{
		PetID:          uuid.New(),
		RequesterID:    previous.UserID,
		TargetMemberID: current.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PreviousOwner.UserID != previous.UserID || out.CurrentOwner.UserID != current.UserID {
		t.Fatalf("unexpected transfer result: %+v", out)
	}
}

func TestGetMyAccessReturnsActiveMember(t *testing.T) {
	t.Parallel()

	member := &ports.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE"}
	svc := newTestService(&fakeMembershipRepo{activeView: member}, &fakeRoleRepo{}, nil)

	out, err := svc.GetMyAccess(context.Background(), member.PetID, member.UserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != member {
		t.Fatalf("unexpected member: %+v", out)
	}
}

func TestLeavePetRemovesNonOwner(t *testing.T) {
	t.Parallel()

	member := &ports.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE", IsPrimaryOwner: false}
	removed := *member
	removed.Status = "REMOVED"
	svc := newTestService(&fakeMembershipRepo{
		activeView:   member,
		removeResult: &removed,
	}, &fakeRoleRepo{}, nil)

	out, err := svc.LeavePet(context.Background(), member.PetID, member.UserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != "REMOVED" {
		t.Fatalf("unexpected removed member: %+v", out)
	}
}

func TestLeavePetRejectsPrimaryOwner(t *testing.T) {
	t.Parallel()

	member := &ports.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE", IsPrimaryOwner: true}
	svc := newTestService(&fakeMembershipRepo{activeView: member}, &fakeRoleRepo{}, nil)

	_, err := svc.LeavePet(context.Background(), member.PetID, member.UserID)
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestCreateCustomRoleRequiresNonEmptyTitle(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{}, nil)

	_, err := svc.CreateCustomRole(context.Background(), CreateCustomRoleParams{PetID: uuid.New(), RequesterID: uuid.New(), Title: "   "})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateCustomRoleReturnsRole(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	role := &ports.RoleView{ID: uuid.New(), PetID: &petID, Kind: "CUSTOM", Title: "Family"}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{createRole: role}, nil)

	out, err := svc.CreateCustomRole(context.Background(), CreateCustomRoleParams{
		PetID:       petID,
		RequesterID: uuid.New(),
		Title:       " Family ",
		Policy:      model.Policy{PetRead: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != role {
		t.Fatalf("unexpected role: %+v", out)
	}
}

func TestUpdateCustomRoleReturnsRole(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	role := &ports.RoleView{ID: uuid.New(), PetID: &petID, Kind: "CUSTOM", Title: "Friends"}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{updateRole: role}, nil)

	out, err := svc.UpdateCustomRole(context.Background(), UpdateCustomRoleParams{
		PetID:       petID,
		RequesterID: uuid.New(),
		RoleID:      role.ID,
		Title:       "Friends",
		Policy:      model.Policy{PetRead: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != role {
		t.Fatalf("unexpected role: %+v", out)
	}
}

func TestDeleteCustomRoleSuccess(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{}, nil)

	if err := svc.DeleteCustomRole(context.Background(), uuid.New(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateInviteRetriesOnConflict(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	roleID := uuid.New()
	inviteRepo := &fakeInviteRepo{createErrSeq: []error{ports.ErrConflict, nil}}
	role := &ports.RoleView{ID: roleID, Kind: "SYSTEM", Code: "VET"}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{byID: role}, inviteRepo)

	res, err := svc.CreateInvite(context.Background(), CreateInviteParams{
		PetID:           petID,
		CreatedByUserID: uuid.New(),
		RoleID:          roleID,
		Policy:          model.Policy{PetRead: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Invite == nil {
		t.Fatalf("expected invite result")
	}
	if inviteRepo.createCalls != 2 {
		t.Fatalf("expected 2 create attempts, got %d", inviteRepo.createCalls)
	}
}

func TestCreateInviteReturnsConflictAfterRetryExhausted(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	roleID := uuid.New()
	inviteRepo := &fakeInviteRepo{createErrSeq: []error{
		ports.ErrConflict, ports.ErrConflict, ports.ErrConflict, ports.ErrConflict, ports.ErrConflict,
	}}
	role := &ports.RoleView{ID: roleID, Kind: "SYSTEM", Code: "VET"}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{byID: role}, inviteRepo)

	_, err := svc.CreateInvite(context.Background(), CreateInviteParams{
		PetID:           petID,
		CreatedByUserID: uuid.New(),
		RoleID:          roleID,
		Policy:          model.Policy{PetRead: true},
	})
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if inviteRepo.createCalls != 5 {
		t.Fatalf("expected 5 create attempts, got %d", inviteRepo.createCalls)
	}
}

func TestCreateInviteRejectsForeignCustomRole(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	anotherPetID := uuid.New()
	roleID := uuid.New()
	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{byID: &ports.RoleView{
		ID:    roleID,
		Kind:  "CUSTOM",
		PetID: &anotherPetID,
	}}, &fakeInviteRepo{})

	_, err := svc.CreateInvite(context.Background(), CreateInviteParams{
		PetID:           petID,
		CreatedByUserID: uuid.New(),
		RoleID:          roleID,
		Policy:          model.Policy{PetRead: true},
	})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateMemberPermissionsBlocksWeakPrimaryOwnerPolicy(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	memberID := uuid.New()
	roleID := uuid.New()
	ownerMember := &ports.MemberView{ID: memberID, PetID: petID, Status: "ACTIVE", IsPrimaryOwner: true}
	svc := newTestService(&fakeMembershipRepo{
		byPetUser:    &ports.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersWrite: true}},
		activeView:   ownerMember,
		updateResult: ownerMember,
	}, &fakeRoleRepo{byID: &ports.RoleView{ID: roleID, Kind: "SYSTEM"}}, nil)

	_, err := svc.UpdateMemberPermissions(context.Background(), UpdateMemberPermissionsParams{
		PetID:       petID,
		RequesterID: uuid.New(),
		MemberID:    memberID,
		RoleID:      roleID,
		Policy:      model.Policy{PetRead: true},
	})
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRemoveMemberMapsRepoConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersWrite: true}},
		removeErr: ports.ErrConflict,
	}, &fakeRoleRepo{}, nil)

	_, err := svc.RemoveMember(context.Background(), RemoveMemberParams{PetID: uuid.New(), RequesterID: uuid.New(), MemberID: uuid.New()})
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRevokeInviteMapsRepoConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{}, &fakeInviteRepo{revokeErr: ports.ErrConflict})

	err := svc.RevokeInvite(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRegenerateInviteLinkRetriesOnConflict(t *testing.T) {
	t.Parallel()

	inviteRepo := &fakeInviteRepo{
		rotateErrSeq: []error{ports.ErrConflict, nil},
		rotateInvite: &ports.InviteView{ID: uuid.New(), PetID: uuid.New(), Status: "ACTIVE"},
	}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &ports.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{}, inviteRepo)

	res, err := svc.RegenerateInviteLink(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Invite == nil {
		t.Fatalf("expected regenerate invite result")
	}
	if res.DeeplinkURL == "" {
		t.Fatalf("expected deeplink url")
	}
	if inviteRepo.rotateCalls != 2 {
		t.Fatalf("expected 2 rotate attempts, got %d", inviteRepo.rotateCalls)
	}
}

func TestRegenerateInviteLinkForbiddenWithoutMembersWrite(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{
			Status: "ACTIVE",
			Policy: model.Policy{MembersWrite: false},
		},
	}, &fakeRoleRepo{}, &fakeInviteRepo{})

	_, err := svc.RegenerateInviteLink(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAcceptInviteByCodeMapsConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, &fakeInviteRepo{
		acceptCodeErr: ports.ErrConflict,
	})
	_, err := svc.AcceptInviteByCode(context.Background(), "123456", uuid.New())
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestAcceptInviteByTokenMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, &fakeInviteRepo{
		acceptTokenErr: ports.ErrNotFound,
	})
	_, err := svc.AcceptInviteByToken(context.Background(), "sometoken", uuid.New())
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetBootstrapReturnsMembersRolesAndPermissions(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	userID := uuid.New()
	member := ports.MemberView{
		ID:     uuid.New(),
		PetID:  petID,
		UserID: userID,
		Status: "ACTIVE",
		Policy: model.Policy{
			MembersRead:  true,
			MembersWrite: true,
		},
	}
	role := ports.RoleView{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"}
	svc := newTestService(&fakeMembershipRepo{
		activeView:  &member,
		activeViews: []ports.MemberView{member},
	}, &fakeRoleRepo{roles: []ports.RoleView{role}}, nil)

	out, err := svc.GetBootstrap(context.Background(), petID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Me.UserID != userID || len(out.Members) != 1 || len(out.Roles) != 1 {
		t.Fatalf("unexpected bootstrap result: %+v", out)
	}
	if !out.CanMembersRead || !out.CanMembersWrite {
		t.Fatalf("unexpected bootstrap permissions: %+v", out)
	}
}

func TestListMembersReturnsActiveMembers(t *testing.T) {
	t.Parallel()

	member := ports.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE"}
	svc := newTestService(&fakeMembershipRepo{
		byPetUser:   &ports.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersRead: true}},
		activeViews: []ports.MemberView{member},
	}, &fakeRoleRepo{}, nil)

	out, err := svc.ListMembers(context.Background(), member.PetID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].ID != member.ID {
		t.Fatalf("unexpected members: %+v", out)
	}
}

func TestListRolesReturnsRoles(t *testing.T) {
	t.Parallel()

	role := ports.RoleView{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"}
	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersRead: true}},
	}, &fakeRoleRepo{roles: []ports.RoleView{role}}, nil)

	out, err := svc.ListRoles(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].ID != role.ID {
		t.Fatalf("unexpected roles: %+v", out)
	}
}

func TestListInvitesSuccess(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersRead: true}},
	}, &fakeRoleRepo{}, nil)

	out, err := svc.ListInvites(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected non-nil invite list")
	}
}

func TestListRolesRequiresMembersReadPermission(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{
			Status: "ACTIVE",
			Policy: model.Policy{MembersRead: false},
		},
	}, &fakeRoleRepo{
		roles: []ports.RoleView{{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"}},
	}, nil)

	_, err := svc.ListRoles(context.Background(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCriticalOwnerPolicy(t *testing.T) {
	t.Parallel()

	weak := model.Policy{
		PetRead: true, PetWrite: true,
		LogRead: true, LogWrite: true,
		HealthRead: true, HealthWrite: true,
		MembersRead: true, MembersWrite: false,
	}
	if criticalOwnerPolicy(weak) {
		t.Fatalf("expected weak policy to be non-critical")
	}

	strong := model.Policy{
		PetRead: true, PetWrite: true,
		LogRead: true, LogWrite: true,
		HealthRead: true, HealthWrite: true,
		MembersRead: true, MembersWrite: true,
	}
	if !criticalOwnerPolicy(strong) {
		t.Fatalf("expected strong policy to be critical")
	}
}

func TestAcceptInviteByCodeSuccess(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)

	out, err := svc.AcceptInviteByCode(context.Background(), "123456", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || out.Member == nil || out.PetID == uuid.Nil {
		t.Fatalf("unexpected accept result: %+v", out)
	}
}

func TestAcceptInviteByTokenSuccess(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)

	out, err := svc.AcceptInviteByToken(context.Background(), "token", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || out.Member == nil || out.PetID == uuid.Nil {
		t.Fatalf("unexpected accept result: %+v", out)
	}
}

func TestAcceptInviteByCodeInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)
	_, err := svc.AcceptInviteByCode(context.Background(), "12", uuid.New())
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAcceptInviteByTokenInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)
	_, err := svc.AcceptInviteByToken(context.Background(), "", uuid.New())
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPreviewInviteByTokenInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil)
	_, err := svc.PreviewInviteByToken(context.Background(), " ")
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPreviewInviteByTokenSuccess(t *testing.T) {
	t.Parallel()

	invite := &ports.InviteView{ID: uuid.New(), PetID: uuid.New(), Status: "ACTIVE"}
	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, &fakeInviteRepo{previewInvite: invite})

	out, err := svc.PreviewInviteByToken(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != invite {
		t.Fatalf("unexpected invite: %+v", out)
	}
}

func TestPreviewInviteByTokenMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, &fakeInviteRepo{
		previewErr: ports.ErrNotFound,
	})
	_, err := svc.PreviewInviteByToken(context.Background(), "token")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetBootstrapRequiresMembersReadPermission(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	svc := newTestService(&fakeMembershipRepo{
		activeView: &ports.MemberView{
			ID:             uuid.New(),
			PetID:          uuid.New(),
			UserID:         uuid.New(),
			Status:         "ACTIVE",
			IsPrimaryOwner: false,
			Role:           ports.RoleView{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"},
			Policy:         model.Policy{MembersRead: false},
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}, &fakeRoleRepo{}, nil)

	_, err := svc.GetBootstrap(context.Background(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListMembersRequiresMembersReadPermission(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &ports.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersRead: false}},
		activeViews: []ports.MemberView{{
			ID:             uuid.New(),
			PetID:          uuid.New(),
			UserID:         uuid.New(),
			Status:         "ACTIVE",
			IsPrimaryOwner: false,
			Role:           ports.RoleView{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"},
			Policy:         model.Policy{},
			CreatedAt:      now,
			UpdatedAt:      now,
		}},
	}, &fakeRoleRepo{}, nil)

	_, err := svc.ListMembers(context.Background(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
