package ports

import (
	"acl/internal/domain/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type InviteCreateInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	CreatedByUserID uuid.UUID
	Status          string
	Token           string
	Code            string
	ExpiresAt       time.Time
	RoleID          uuid.UUID
	Policy          model.Policy
}

type InviteView struct {
	ID               uuid.UUID
	PetID            uuid.UUID
	Status           string
	Token            string
	Code             string
	ExpiresAt        time.Time
	Role             RoleView
	Policy           model.Policy
	CreatedByUserID  uuid.UUID
	CreatedAt        time.Time
	ConsumedAt       *time.Time
	ConsumedByUserID *uuid.UUID
}

type InviteRepository interface {
	Create(ctx context.Context, in InviteCreateInput) (*InviteView, error)
	ListActiveByPet(ctx context.Context, petID uuid.UUID) ([]InviteView, error)
	GetActiveByToken(ctx context.Context, token string) (*InviteView, error)
	AcceptByCode(ctx context.Context, code string, acceptedByUserID uuid.UUID) (*MemberView, uuid.UUID, error)
	AcceptByToken(ctx context.Context, token string, acceptedByUserID uuid.UUID) (*MemberView, uuid.UUID, error)
	RotateTokenByID(ctx context.Context, petID, inviteID uuid.UUID, token string) (*InviteView, error)
	RevokeByID(ctx context.Context, petID, inviteID uuid.UUID) error
}
