package usecase

import (
	"context"
	"profile/internal/model"

	"github.com/google/uuid"
)

type UpdateProfileInfoUseCase struct{ deps *dependencies }

type UpdateProfileInfoInput struct {
	FirstName *string
	LastName  *string
}

func (uc *UpdateProfileInfoUseCase) Execute(ctx context.Context, userID uuid.UUID, in UpdateProfileInfoInput) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	profile, err := uc.deps.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.FirstName != nil {
		profile.FirstName = normalizeOptionalString(in.FirstName)
	}
	if in.LastName != nil {
		profile.LastName = normalizeOptionalString(in.LastName)
	}

	if err := uc.deps.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

type UpdatePreferencesUseCase struct{ deps *dependencies }

type UpdatePreferencesInput struct {
	Locale   *string
	Timezone *string
}

func (uc *UpdatePreferencesUseCase) Execute(ctx context.Context, userID uuid.UUID, in UpdatePreferencesInput) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	profile, err := uc.deps.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Locale != nil {
		locale, err := uc.deps.normalizeLocale(in.Locale)
		if err != nil {
			return nil, err
		}
		profile.Locale = locale
	}
	if in.Timezone != nil {
		timezone, err := uc.deps.normalizeTimezone(in.Timezone)
		if err != nil {
			return nil, err
		}
		profile.Timezone = timezone
	}

	if err := uc.deps.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
