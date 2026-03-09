package service

import (
	"acl/internal/model"
	"acl/internal/repository"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeMembershipRepo struct {
	byPetUser       *repository.MembershipAccess
	activeByPetUser *repository.MembershipAccess
	activeView      *repository.MemberView
	activeViews     []repository.MemberView
	ownerResult     *repository.MemberView
	ownerErr        error
	updateResult    *repository.MemberView
	removeResult    *repository.MemberView
	updateErr       error
	removeErr       error
	err             error
}

func (f *fakeMembershipRepo) GetByPetAndUser(_ context.Context, _, _ uuid.UUID) (*repository.MembershipAccess, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.byPetUser == nil {
		return nil, repository.ErrNotFound
	}
	return f.byPetUser, nil
}

func (f *fakeMembershipRepo) GetActiveByPetAndUser(_ context.Context, _, _ uuid.UUID) (*repository.MembershipAccess, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.activeByPetUser == nil {
		return nil, repository.ErrNotFound
	}
	return f.activeByPetUser, nil
}

func (f *fakeMembershipRepo) CreateOwner(_ context.Context, _, _ uuid.UUID, _ model.Policy) (*repository.MemberView, error) {
	if f.ownerErr != nil {
		return nil, f.ownerErr
	}
	if f.ownerResult == nil {
		return nil, repository.ErrNotFound
	}
	return f.ownerResult, nil
}

func (f *fakeMembershipRepo) GetActiveViewByPetAndUser(_ context.Context, _, _ uuid.UUID) (*repository.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.activeView == nil {
		return nil, repository.ErrNotFound
	}
	return f.activeView, nil
}

func (f *fakeMembershipRepo) GetByIDAndPet(_ context.Context, _, _ uuid.UUID) (*repository.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.activeView == nil {
		return nil, repository.ErrNotFound
	}
	return f.activeView, nil
}

func (f *fakeMembershipRepo) ListActiveViewsByPet(_ context.Context, _ uuid.UUID) ([]repository.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.activeViews, nil
}

func (f *fakeMembershipRepo) ListActiveViewsByUser(_ context.Context, _ uuid.UUID) ([]repository.MemberView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.activeViews, nil
}

func (f *fakeMembershipRepo) UpdatePermissions(_ context.Context, _, _ uuid.UUID, _ uuid.UUID, _ model.Policy, _ *uuid.UUID) (*repository.MemberView, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updateResult == nil {
		return nil, repository.ErrNotFound
	}
	return f.updateResult, nil
}

func (f *fakeMembershipRepo) RemoveMember(_ context.Context, _, _ uuid.UUID, _ uuid.UUID) (*repository.MemberView, error) {
	if f.removeErr != nil {
		return nil, f.removeErr
	}
	if f.removeResult == nil {
		return nil, repository.ErrNotFound
	}
	return f.removeResult, nil
}

type fakeRoleRepo struct {
	roles      []repository.RoleView
	byID       *repository.RoleView
	createRole *repository.RoleView
	createErr  error
	deleteErr  error
	err        error
}

func (f *fakeRoleRepo) GetByID(_ context.Context, _ uuid.UUID) (*repository.RoleView, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.byID == nil {
		return nil, repository.ErrNotFound
	}
	return f.byID, nil
}

func (f *fakeRoleRepo) ListSystemAndPetRoles(_ context.Context, _ uuid.UUID) ([]repository.RoleView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.roles, nil
}

func (f *fakeRoleRepo) CreateCustom(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) (*repository.RoleView, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createRole == nil {
		return nil, repository.ErrNotFound
	}
	return f.createRole, nil
}

func (f *fakeRoleRepo) DeleteCustomIfUnused(_ context.Context, _, _ uuid.UUID) error {
	return f.deleteErr
}

type fakePresetRepo struct {
	items  []repository.PermissionPresetView
	exists bool
	err    error
}

func (f *fakePresetRepo) ListSystem(_ context.Context) ([]repository.PermissionPresetView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakePresetRepo) ExistsByID(_ context.Context, _ uuid.UUID) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.exists, nil
}

type fakeInviteRepo struct {
	createErrSeq   []error
	createCalls    int
	createInvite   *repository.InviteView
	previewInvite  *repository.InviteView
	previewErr     error
	revokeErr      error
	acceptCodeErr  error
	acceptTokenErr error
}

func (f *fakeInviteRepo) Create(_ context.Context, _ repository.InviteCreateInput) (*repository.InviteView, error) {
	call := f.createCalls
	f.createCalls++
	if call < len(f.createErrSeq) && f.createErrSeq[call] != nil {
		return nil, f.createErrSeq[call]
	}
	if f.createInvite == nil {
		return &repository.InviteView{ID: uuid.New(), PetID: uuid.New(), Status: "ACTIVE"}, nil
	}
	return f.createInvite, nil
}

func (f *fakeInviteRepo) ListActiveByPet(_ context.Context, _ uuid.UUID) ([]repository.InviteView, error) {
	return []repository.InviteView{}, nil
}

func (f *fakeInviteRepo) GetActiveByTokenHash(_ context.Context, _ string) (*repository.InviteView, error) {
	if f.previewErr != nil {
		return nil, f.previewErr
	}
	if f.previewInvite == nil {
		return nil, repository.ErrNotFound
	}
	return f.previewInvite, nil
}

func (f *fakeInviteRepo) AcceptByCode(_ context.Context, _ string, _ uuid.UUID) (*repository.MemberView, uuid.UUID, error) {
	if f.acceptCodeErr != nil {
		return nil, uuid.Nil, f.acceptCodeErr
	}
	member := &repository.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE"}
	return member, member.PetID, nil
}

func (f *fakeInviteRepo) AcceptByTokenHash(_ context.Context, _ string, _ uuid.UUID) (*repository.MemberView, uuid.UUID, error) {
	if f.acceptTokenErr != nil {
		return nil, uuid.Nil, f.acceptTokenErr
	}
	member := &repository.MemberView{ID: uuid.New(), PetID: uuid.New(), UserID: uuid.New(), Status: "ACTIVE"}
	return member, member.PetID, nil
}

func (f *fakeInviteRepo) RevokeByID(_ context.Context, _, _ uuid.UUID) error {
	return f.revokeErr
}

func newTestService(m *fakeMembershipRepo, r *fakeRoleRepo, p *fakePresetRepo, i *fakeInviteRepo) *ACLService {
	if p == nil {
		p = &fakePresetRepo{exists: true}
	}
	if i == nil {
		i = &fakeInviteRepo{}
	}
	return New(m, r, p, i, Options{InviteDeeplinkBase: "myapp://invite?token="})
}

func TestListPresetsReturnsSystemPresets(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, &fakePresetRepo{
		items: []repository.PermissionPresetView{{
			ID:       uuid.New(),
			Name:     "Owner Full Access",
			RoleCode: "OWNER",
			Policy:   model.Policy{MembersWrite: true},
		}},
		exists: true,
	}, nil)

	items, err := svc.ListPresets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(items))
	}
}

func TestCheckAllowed(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		MemberID:       uuid.New(),
		Status:         "ACTIVE",
		IsPrimaryOwner: false,
		Policy:         model.Policy{PetRead: true},
	}}, &fakeRoleRepo{}, nil, nil)

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

	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{Status: "ACTIVE"}}, &fakeRoleRepo{}, nil, nil)

	_, err := svc.Check(context.Background(), CheckParams{PetID: uuid.New(), UserID: uuid.New(), Action: "unknown_action"})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateOwnerMembershipMapsConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		ownerErr: repository.ErrConflict,
	}, &fakeRoleRepo{}, nil, nil)

	_, err := svc.CreateOwnerMembership(context.Background(), uuid.New(), uuid.New())
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestCreateCustomRoleRequiresNonEmptyTitle(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{}, nil, nil)

	_, err := svc.CreateCustomRole(context.Background(), CreateCustomRoleParams{PetID: uuid.New(), RequesterID: uuid.New(), Title: "   "})
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateInviteRetriesOnConflict(t *testing.T) {
	t.Parallel()

	petID := uuid.New()
	roleID := uuid.New()
	inviteRepo := &fakeInviteRepo{createErrSeq: []error{repository.ErrConflict, nil}}
	role := &repository.RoleView{ID: roleID, Kind: "SYSTEM", Code: "VET"}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{byID: role}, &fakePresetRepo{exists: true}, inviteRepo)

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
		repository.ErrConflict, repository.ErrConflict, repository.ErrConflict, repository.ErrConflict, repository.ErrConflict,
	}}
	role := &repository.RoleView{ID: roleID, Kind: "SYSTEM", Code: "VET"}
	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{byID: role}, &fakePresetRepo{exists: true}, inviteRepo)

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
	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{byID: &repository.RoleView{
		ID:    roleID,
		Kind:  "CUSTOM",
		PetID: &anotherPetID,
	}}, &fakePresetRepo{exists: true}, &fakeInviteRepo{})

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
	ownerMember := &repository.MemberView{ID: memberID, PetID: petID, Status: "ACTIVE", IsPrimaryOwner: true}
	svc := newTestService(&fakeMembershipRepo{
		byPetUser:    &repository.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersWrite: true}},
		activeView:   ownerMember,
		updateResult: ownerMember,
	}, &fakeRoleRepo{byID: &repository.RoleView{ID: roleID, Kind: "SYSTEM"}}, &fakePresetRepo{exists: true}, nil)

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
		byPetUser: &repository.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersWrite: true}},
		removeErr: repository.ErrConflict,
	}, &fakeRoleRepo{}, nil, nil)

	_, err := svc.RemoveMember(context.Background(), RemoveMemberParams{PetID: uuid.New(), RequesterID: uuid.New(), MemberID: uuid.New()})
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestRevokeInviteMapsRepoConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		Status: "ACTIVE",
		Policy: model.Policy{MembersWrite: true},
	}}, &fakeRoleRepo{}, nil, &fakeInviteRepo{revokeErr: repository.ErrConflict})

	err := svc.RevokeInvite(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestAcceptInviteByCodeMapsConflict(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil, &fakeInviteRepo{
		acceptCodeErr: repository.ErrConflict,
	})
	_, err := svc.AcceptInviteByCode(context.Background(), "123456", uuid.New())
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestAcceptInviteByTokenMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil, &fakeInviteRepo{
		acceptTokenErr: repository.ErrNotFound,
	})
	_, err := svc.AcceptInviteByToken(context.Background(), "sometoken", uuid.New())
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListRolesRequiresMembersReadPermission(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &repository.MembershipAccess{
			Status: "ACTIVE",
			Policy: model.Policy{MembersRead: false},
		},
	}, &fakeRoleRepo{
		roles: []repository.RoleView{{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"}},
	}, nil, nil)

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
		TaskRead: true, TaskWrite: true,
		MembersRead: true, MembersWrite: false,
	}
	if criticalOwnerPolicy(weak) {
		t.Fatalf("expected weak policy to be non-critical")
	}

	strong := model.Policy{
		PetRead: true, PetWrite: true,
		LogRead: true, LogWrite: true,
		HealthRead: true, HealthWrite: true,
		TaskRead: true, TaskWrite: true,
		MembersRead: true, MembersWrite: true,
	}
	if !criticalOwnerPolicy(strong) {
		t.Fatalf("expected strong policy to be critical")
	}
}

func TestAcceptInviteByCodeInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil, nil)
	_, err := svc.AcceptInviteByCode(context.Background(), "12", uuid.New())
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAcceptInviteByTokenInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil, nil)
	_, err := svc.AcceptInviteByToken(context.Background(), "", uuid.New())
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPreviewInviteByTokenInvalidInput(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil, nil)
	_, err := svc.PreviewInviteByToken(context.Background(), " ")
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPreviewInviteByTokenMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(&fakeMembershipRepo{}, &fakeRoleRepo{}, nil, &fakeInviteRepo{
		previewErr: repository.ErrNotFound,
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
		activeView: &repository.MemberView{
			ID:             uuid.New(),
			PetID:          uuid.New(),
			UserID:         uuid.New(),
			Status:         "ACTIVE",
			IsPrimaryOwner: false,
			Role:           repository.RoleView{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"},
			Policy:         model.Policy{MembersRead: false},
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}, &fakeRoleRepo{}, nil, nil)

	_, err := svc.GetBootstrap(context.Background(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListMembersRequiresMembersReadPermission(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	svc := newTestService(&fakeMembershipRepo{
		byPetUser: &repository.MembershipAccess{Status: "ACTIVE", Policy: model.Policy{MembersRead: false}},
		activeViews: []repository.MemberView{{
			ID:             uuid.New(),
			PetID:          uuid.New(),
			UserID:         uuid.New(),
			Status:         "ACTIVE",
			IsPrimaryOwner: false,
			Role:           repository.RoleView{ID: uuid.New(), Kind: "SYSTEM", Code: "OWNER", Title: "Owner"},
			Policy:         model.Policy{},
			CreatedAt:      now,
			UpdatedAt:      now,
		}},
	}, &fakeRoleRepo{}, nil, nil)

	_, err := svc.ListMembers(context.Background(), uuid.New(), uuid.New())
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
