package usecase

import (
	"context"
	"profile/internal/model"

	"github.com/google/uuid"
)

type GetProfileUseCase struct{ deps *dependencies }

func (uc *GetProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	return uc.deps.profiles.GetByUserID(ctx, userID)
}

type DeleteProfileUseCase struct{ deps *dependencies }

func (uc *DeleteProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidInput
	}
	return uc.deps.profiles.Delete(ctx, userID)
}
