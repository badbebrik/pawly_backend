package service

import (
	"auth/internal/infrastructure/rabbit"
	"auth/internal/repository"
	"auth/internal/security"
	"auth/internal/transport/http/middleware"
	"auth/internal/verification"
	"context"
	"errors"

	"github.com/google/uuid"
)

type EmailVerificationMeta struct {
	Channel            string
	CodeTTLSeconds     int
	CanResendInSeconds int
}

type ResendEmailVerificationInput struct {
	Email string
}

type ResendEmailVerificationOutput struct {
	UserID       uuid.UUID
	Verification EmailVerificationMeta
}

func (s *Service) ResendEmailVerification(ctx context.Context, in ResendEmailVerificationInput) (*ResendEmailVerificationOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) {
		return nil, ErrIncorrectFormat
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if !u.IsActive {
		return nil, ErrUserBlocked
	}
	if u.IsVerified {
		return nil, ErrEmailAlreadyVerified
	}

	meta, err := s.sendRegistrationVerification(ctx, u.ID, email, "", "")
	out := &ResendEmailVerificationOutput{
		UserID:       u.ID,
		Verification: meta,
	}
	if err != nil {
		return out, err
	}

	return out, nil
}

func (s *Service) sendRegistrationVerification(ctx context.Context, userID uuid.UUID, email, firstName, lastName string) (EmailVerificationMeta, error) {
	loc := middleware.LocaleFromCtx(ctx, "ru")

	code, ttlSeconds, resendInSeconds, err := s.verification.RequestCode(ctx, email, "registration")
	meta := EmailVerificationMeta{
		Channel:            "email",
		CodeTTLSeconds:     ttlSeconds,
		CanResendInSeconds: resendInSeconds,
	}
	if err != nil {
		if errors.Is(err, verification.ErrResendTooSoon) {
			return meta, ErrCannotResendYet
		}
		return meta, ErrVerificationFailed
	}

	if err := s.notifier.SendEmailVerification(ctx, rabbit.EmailVerificationPayload{
		UserID:     userID,
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		Code:       code,
		TTLSeconds: ttlSeconds,
		Locale:     loc,
	}); err != nil {
		return meta, ErrVerificationFailed
	}

	return meta, nil
}
