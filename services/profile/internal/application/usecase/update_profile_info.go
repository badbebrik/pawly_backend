package usecase

import (
	"context"
	"profile/internal/domain/model"

	"github.com/google/uuid"
)

type UpdateProfileInfo struct{ deps *dependencies }

type UpdateProfileInfoParams struct {
	FirstName *string
	LastName  *string
}

func (uc *UpdateProfileInfo) Execute(ctx context.Context, userID uuid.UUID, in UpdateProfileInfoParams) (*model.Profile, error) {
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

type UpdatePreferences struct{ deps *dependencies }

type UpdatePreferencesParams struct {
	Locale   *string
	Timezone *string
}

func (uc *UpdatePreferences) Execute(ctx context.Context, userID uuid.UUID, in UpdatePreferencesParams) (*model.Profile, error) {
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
