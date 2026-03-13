package service

import (
	"auth/internal/domain/model"
	"auth/internal/repository"
	"auth/internal/security"
	"auth/internal/transport/http/middleware"
	"context"
	"errors"
	"github.com/google/uuid"
)

type RegisterEmailInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterEmailOutput struct {
	UserID       uuid.UUID
	Verification struct {
		Channel            string
		CodeTTLSeconds     int
		CanResendInSeconds int
	}
}

func (s *Service) RegisterEmail(ctx context.Context, in RegisterEmailInput) (*RegisterEmailOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) {
		return nil, ErrIncorrectFormat
	}
	if !security.ValidatePassword(in.Password) {
		return nil, ErrWeakPassword
	}

	existing, err := s.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	loc := middleware.LocaleFromCtx(ctx, "ru")

	if existing != nil {
		if existing.IsVerified {
			return nil, ErrEmailAlreadyTaken
		}
		if !existing.IsActive {
			return nil, ErrUserBlocked
		}

		hash, err := security.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}
		if err := s.users.UpdatePasswordHash(ctx, existing.ID, hash); err != nil {
			return nil, err
		}
		if err := s.profile.CreateProfile(ctx, existing.ID, loc, in.FirstName, in.LastName); err != nil {
			return nil, ErrProfileCreationFailed
		}

		meta, verr := s.sendRegistrationVerification(ctx, existing.ID, email, in.FirstName, in.LastName)
		out := &RegisterEmailOutput{
			UserID: existing.ID,
		}
		out.Verification.Channel = meta.Channel
		out.Verification.CodeTTLSeconds = meta.CodeTTLSeconds
		out.Verification.CanResendInSeconds = meta.CanResendInSeconds
		if verr != nil {
			return out, verr
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

	if err := s.createUserWithProfile(ctx, user, loc, in.FirstName, in.LastName); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, ErrEmailAlreadyTaken
		}
		return nil, err
	}

	meta, verr := s.sendRegistrationVerification(ctx, user.ID, email, in.FirstName, in.LastName)
	out := &RegisterEmailOutput{
		UserID: user.ID,
	}
	out.Verification.Channel = meta.Channel
	out.Verification.CodeTTLSeconds = meta.CodeTTLSeconds
	out.Verification.CanResendInSeconds = meta.CanResendInSeconds

	if verr != nil {
		return out, verr
	}

	return out, nil
}
