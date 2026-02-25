package repository

import (
	"acl/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type MembershipAccess struct {
	MemberID       uuid.UUID
	Status         string
	IsPrimaryOwner bool
	Policy         model.Policy
}

type RoleView struct {
	ID              uuid.UUID
	Kind            string
	PetID           *uuid.UUID
	Code            string
	Title           string
	CreatedByUserID *uuid.UUID
}

type MemberView struct {
	ID             uuid.UUID
	PetID          uuid.UUID
	UserID         uuid.UUID
	Status         string
	IsPrimaryOwner bool
	Role           RoleView
	Policy         model.Policy
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MembershipRepository interface {
	GetByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*MembershipAccess, error)
	GetActiveByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*MembershipAccess, error)
	CreateOwner(ctx context.Context, petID, ownerUserID uuid.UUID, policy model.Policy) (*MemberView, error)
	GetActiveViewByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*MemberView, error)
	GetByIDAndPet(ctx context.Context, petID, memberID uuid.UUID) (*MemberView, error)
	ListActiveViewsByPet(ctx context.Context, petID uuid.UUID) ([]MemberView, error)
	ListActivePetIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	UpdatePermissions(ctx context.Context, petID, memberID uuid.UUID, roleID uuid.UUID, policy model.Policy, basePresetID *uuid.UUID) (*MemberView, error)
	RemoveMember(ctx context.Context, petID, memberID, removedByUserID uuid.UUID) (*MemberView, error)
}
