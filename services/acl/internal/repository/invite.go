package repository

import (
	"acl/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type InviteCreateInput struct {
	ID              uuid.UUID
	PetID           uuid.UUID
	CreatedByUserID uuid.UUID
	Status          string
	TokenHash       string
	TokenValue      string
	Code            string
	ExpiresAt       time.Time
	RoleID          uuid.UUID
	Policy          model.Policy
	BasePresetID    *uuid.UUID
}

type InviteView struct {
	ID               uuid.UUID
	PetID            uuid.UUID
	Status           string
	TokenValue       string
	Code             string
	ExpiresAt        time.Time
	Role             RoleView
	Policy           model.Policy
	BasePresetID     *uuid.UUID
	CreatedByUserID  uuid.UUID
	CreatedAt        time.Time
	ConsumedAt       *time.Time
	ConsumedByUserID *uuid.UUID
}

type InviteRepository interface {
	Create(ctx context.Context, in InviteCreateInput) (*InviteView, error)
	ListActiveByPet(ctx context.Context, petID uuid.UUID) ([]InviteView, error)
	GetActiveByTokenHash(ctx context.Context, tokenHash string) (*InviteView, error)
	AcceptByCode(ctx context.Context, code string, acceptedByUserID uuid.UUID) (*MemberView, uuid.UUID, error)
	AcceptByTokenHash(ctx context.Context, tokenHash string, acceptedByUserID uuid.UUID) (*MemberView, uuid.UUID, error)
	RotateTokenHashByID(ctx context.Context, petID, inviteID uuid.UUID, tokenHash, tokenValue string) (*InviteView, error)
	RevokeByID(ctx context.Context, petID, inviteID uuid.UUID) error
}
