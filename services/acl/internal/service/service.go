package service

import (
	"acl/internal/model"
	"context"

	"github.com/google/uuid"
)

type ACLService struct{}

func New() *ACLService {
	return &ACLService{}
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

func (s *ACLService) Check(_ context.Context, _ CheckParams) (bool, error) {
	return false, ErrNotImplemented
}

func (s *ACLService) IsMember(_ context.Context, _ uuid.UUID, _ uuid.UUID) (bool, error) {
	return false, ErrNotImplemented
}

func (s *ACLService) GetPolicy(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*PolicyResult, error) {
	return nil, ErrNotImplemented
}

func (s *ACLService) ListPetsForUser(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, ErrNotImplemented
}
