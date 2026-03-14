package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"

	"github.com/google/uuid"
)

type RegisterEmailUseCase struct {
	deps *dependencies
}

type RegisterEmailInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Locale    string
	Timezone  string
}

type RegisterEmailOutput struct {
	UserID       uuid.UUID
	Verification struct {
		Channel            string
		CodeTTLSeconds     int
		CanResendInSeconds int
	}
}

func (uc *RegisterEmailUseCase) Execute(ctx context.Context, in RegisterEmailInput) (*RegisterEmailOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) {
		return nil, ErrIncorrectFormat
	}
	if !security.ValidatePassword(in.Password) {
		return nil, ErrWeakPassword
	}

	existing, err := uc.deps.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return nil, err
	}

	locale := normalizeLocale(in.Locale)

	if existing != nil {
		if existing.IsVerified {
			return nil, ErrEmailAlreadyTaken
		}
		if err := existing.RequireActive(); err != nil {
			return nil, ErrUserBlocked
		}

		hash, err := security.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}
		if err := uc.deps.users.UpdatePasswordHash(ctx, existing.ID, hash); err != nil {
			return nil, err
		}
		meta, sendErr := sendRegistrationVerification(ctx, uc.deps, existing.ID, email, in.FirstName, in.LastName, locale)
		out := &RegisterEmailOutput{UserID: existing.ID}
		out.Verification.Channel = meta.Channel
		out.Verification.CodeTTLSeconds = meta.CodeTTLSeconds
		out.Verification.CanResendInSeconds = meta.CanResendInSeconds
		if sendErr != nil {
			return out, sendErr
		}
		return out, nil
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: &hash,
		IsVerified:   false,
		IsActive:     true,
	}

	if err := uc.deps.createUserWithProfile(ctx, user, locale, normalizeTimezone(in.Timezone), in.FirstName, in.LastName); err != nil {
		if errors.Is(err, ports.ErrConflict) {
			return nil, ErrEmailAlreadyTaken
		}
		return nil, err
	}

	meta, sendErr := sendRegistrationVerification(ctx, uc.deps, user.ID, email, in.FirstName, in.LastName, locale)
	out := &RegisterEmailOutput{UserID: user.ID}
	out.Verification.Channel = meta.Channel
	out.Verification.CodeTTLSeconds = meta.CodeTTLSeconds
	out.Verification.CanResendInSeconds = meta.CanResendInSeconds

	if sendErr != nil {
		return out, sendErr
	}

	return out, nil
}
