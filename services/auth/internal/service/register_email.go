package service

import (
	"auth/internal/domain/model"
	"auth/internal/infrastructure/rabbit"
	"auth/internal/repository"
	"auth/internal/security"
	"auth/internal/transport/http/middleware"
	"auth/internal/verification"
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

	var user *model.User

	if existing != nil {
		return nil, ErrEmailAlreadyTaken
	}

	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user = &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: &hash,
		IsVerified:   false,
		IsActive:     true,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, ErrEmailAlreadyTaken
	}

	if err := s.profile.CreateProfile(ctx, user.ID, loc); err != nil {
		_ = s.users.Delete(ctx, user.ID)
		return nil, ErrProfileCreationFailed
	}

	code, ttlSeconds, resendInSeconds, verr := s.verification.RequestCode(ctx, email, "registration")
	out := &RegisterEmailOutput{
		UserID: user.ID,
	}
	out.Verification.Channel = "email"
	out.Verification.CodeTTLSeconds = ttlSeconds
	out.Verification.CanResendInSeconds = resendInSeconds

	if verr != nil {
		if errors.Is(verr, verification.ErrResendTooSoon) {
			return out, ErrCannotResendYet
		}
		return nil, ErrVerificationFailed
	}

	if err := s.notifier.SendEmailVerification(ctx, rabbit.EmailVerificationPayload{
		UserID:     user.ID,
		Email:      email,
		FirstName:  in.FirstName,
		LastName:   in.LastName,
		Code:       code,
		TTLSeconds: ttlSeconds,
		Locale:     loc,
	}); err != nil {
		return nil, ErrVerificationFailed
	}

	return out, nil
}
