package usecase

import (
	"context"
	"profile/internal/domain/model"

	"github.com/google/uuid"
)

type GetProfile struct{ deps *dependencies }

func (uc *GetProfile) Execute(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return uc.deps.profiles.GetByUserID(ctx, userID)
}

type DeleteProfile struct{ deps *dependencies }

func (uc *DeleteProfile) Execute(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidInput
	}
	return uc.deps.profiles.Delete(ctx, userID)
}
