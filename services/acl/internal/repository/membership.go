package repository

import (
	"acl/internal/model"
	"context"

	"github.com/google/uuid"
)

type MembershipAccess struct {
	MemberID       uuid.UUID
	Status         string
	IsPrimaryOwner bool
	Policy         model.Policy
}

type MembershipRepository interface {
	GetByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*MembershipAccess, error)
	GetActiveByPetAndUser(ctx context.Context, petID, userID uuid.UUID) (*MembershipAccess, error)
	ListActivePetIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
