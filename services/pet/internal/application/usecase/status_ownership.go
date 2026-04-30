package usecase

import (
	"context"
	"pet/internal/application/ports"
	"pet/internal/domain/model"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Pet) ChangePetStatus(ctx context.Context, p ChangePetStatusParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetWrite)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	status := strings.ToUpper(strings.TrimSpace(p.Status))
	var (
		missingSince *time.Time
		archivedAt   *time.Time
	)
	switch status {
	case "ACTIVE":
	case "MISSING":
		if p.MissingSince == nil {
			return nil, ErrInvalidInput
		}
		missingSince = p.MissingSince
	case "ARCHIVED":
		now := time.Now().UTC()
		archivedAt = &now
	default:
		return nil, ErrInvalidInput
	}

	pet, err := s.repo.UpdateStatus(ctx, p.PetID, p.RowVersion, status, missingSince, archivedAt)
	if err != nil {
		switch err {
		case ports.ErrNotFound:
			return nil, ErrNotFound
		case ports.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}

	return pet, nil
}

func (s *Pet) TransferOwnership(ctx context.Context, p TransferPetOwnershipParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 || p.TargetMemberID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	current, err := s.repo.GetByID(ctx, p.PetID)
	if err != nil {
		if err == ports.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if current.OwnerUserID != p.UserID {
		return nil, ErrForbidden
	}

	transferRes, err := s.acl.TransferOwnership(ctx, p.PetID, p.UserID, p.TargetMemberID)
	if err != nil {
		switch err {
		case ErrForbidden:
			return nil, ErrForbidden
		case ErrNotFound:
			return nil, ErrNotFound
		case ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}

	updated, err := s.repo.UpdateOwner(ctx, p.PetID, p.RowVersion, transferRes.CurrentOwnerUserID)
	if err == nil {
		return updated, nil
	}

	rollbackRes, rollbackErr := s.acl.TransferOwnership(ctx, p.PetID, transferRes.CurrentOwnerUserID, transferRes.PreviousOwnerMemberID)
	if rollbackErr != nil {
		return nil, rollbackErr
	}
	if rollbackRes.CurrentOwnerUserID != p.UserID {
		return nil, ErrConflict
	}

	switch err {
	case ports.ErrNotFound:
		return nil, ErrNotFound
	case ports.ErrConflict:
		return nil, ErrConflict
	default:
		return nil, err
	}
}
