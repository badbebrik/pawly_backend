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
	petIDs          []uuid.UUID
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

func (f *fakeMembershipRepo) GetActiveViewByPetAndUser(_ context.Context, _, _ uuid.UUID) (*repository.MemberView, error) {
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

func (f *fakeMembershipRepo) ListActivePetIDsByUser(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.petIDs, nil
}

type fakeRoleRepo struct {
	roles []repository.RoleView
	err   error
}

func (f *fakeRoleRepo) ListSystemAndPetRoles(_ context.Context, _ uuid.UUID) ([]repository.RoleView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.roles, nil
}

func TestCheckAllowed(t *testing.T) {
	t.Parallel()

	svc := New(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		MemberID:       uuid.New(),
		Status:         "ACTIVE",
		IsPrimaryOwner: false,
		Policy:         model.Policy{PetRead: true},
	}}, &fakeRoleRepo{})

	allowed, err := svc.Check(context.Background(), CheckParams{
		PetID:  uuid.New(),
		UserID: uuid.New(),
		Action: string(ActionPetRead),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected allowed=true")
	}
}

func TestCheckInvalidAction(t *testing.T) {
	t.Parallel()

	svc := New(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		MemberID:       uuid.New(),
		Status:         "ACTIVE",
		IsPrimaryOwner: false,
		Policy:         model.Policy{},
	}}, &fakeRoleRepo{})

	_, err := svc.Check(context.Background(), CheckParams{
		PetID:  uuid.New(),
		UserID: uuid.New(),
		Action: "unknown_action",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCheckRemovedMemberForbidden(t *testing.T) {
	t.Parallel()

	svc := New(&fakeMembershipRepo{byPetUser: &repository.MembershipAccess{
		MemberID:       uuid.New(),
		Status:         "REMOVED",
		IsPrimaryOwner: false,
		Policy:         model.Policy{PetRead: true},
	}}, &fakeRoleRepo{})

	_, err := svc.Check(context.Background(), CheckParams{
		PetID:  uuid.New(),
		UserID: uuid.New(),
		Action: string(ActionPetRead),
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListMembersRequiresMembersViewPermission(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	svc := New(&fakeMembershipRepo{
		byPetUser: &repository.MembershipAccess{
			MemberID:       uuid.New(),
			Status:         "ACTIVE",
			IsPrimaryOwner: false,
			Policy:         model.Policy{MembersView: false},
		},
		activeViews: []repository.MemberView{{
			ID:             uuid.New(),
			PetID:          uuid.New(),
			UserID:         uuid.New(),
			Status:         "ACTIVE",
			IsPrimaryOwner: false,
			Role: repository.RoleView{
				ID:    uuid.New(),
				Kind:  "SYSTEM",
				Code:  "OWNER",
				Title: "Owner",
			},
			Policy:    model.Policy{},
			CreatedAt: now,
			UpdatedAt: now,
		}},
	}, &fakeRoleRepo{})

	_, err := svc.ListMembers(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
