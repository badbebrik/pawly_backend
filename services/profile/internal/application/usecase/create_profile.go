package usecase

import (
	"context"
	"profile/internal/domain/model"

	"github.com/google/uuid"
)

type CreateProfile struct {
	deps *dependencies
}

type CreateProfileParams struct {
	UserID    uuid.UUID
	Locale    *string
	Timezone  *string
	FirstName *string
	LastName  *string
}

func (uc *CreateProfile) Execute(ctx context.Context, in CreateProfileParams) (*model.Profile, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	locale, err := uc.deps.normalizeLocale(in.Locale)
	if err != nil {
		return nil, err
	}
	timezone, err := uc.deps.normalizeTimezone(in.Timezone)
	if err != nil {
		return nil, err
	}

	profile := &model.Profile{
		UserID:    in.UserID,
		FirstName: normalizeOptionalString(in.FirstName),
		LastName:  normalizeOptionalString(in.LastName),
		Locale:    locale,
		Timezone:  timezone,
	}

	if err := uc.deps.profiles.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
