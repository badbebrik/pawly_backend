package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/security"
	"context"
	"errors"

	"github.com/google/uuid"
)

type EmailVerificationMeta struct {
	Channel            string
	CodeTTLSeconds     int
	CanResendInSeconds int
}

type ResendEmailVerificationUseCase struct {
	deps *dependencies
}

type ResendEmailVerificationInput struct {
	Email  string
	Locale string
}

type ResendEmailVerificationOutput struct {
	UserID       uuid.UUID
	Verification EmailVerificationMeta
}

func (uc *ResendEmailVerificationUseCase) Execute(ctx context.Context, in ResendEmailVerificationInput) (*ResendEmailVerificationOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) {
		return nil, ErrIncorrectFormat
	}

	user, err := uc.deps.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if err := user.RequireActive(); err != nil {
		return nil, ErrUserBlocked
	}
	if user.IsVerified {
		return nil, ErrEmailAlreadyVerified
	}

	meta, sendErr := sendRegistrationVerification(ctx, uc.deps, user.ID, email, "", "", in.Locale)
	out := &ResendEmailVerificationOutput{
		UserID:       user.ID,
		Verification: meta,
	}
	if sendErr != nil {
		return out, sendErr
	}

	return out, nil
}

func sendRegistrationVerification(ctx context.Context, deps *dependencies, userID uuid.UUID, email, firstName, lastName, locale string) (EmailVerificationMeta, error) {
	code, ttlSeconds, resendInSeconds, err := deps.verification.RequestCode(ctx, email, "registration")
	meta := EmailVerificationMeta{
		Channel:            "email",
		CodeTTLSeconds:     ttlSeconds,
		CanResendInSeconds: resendInSeconds,
	}
	if err != nil {
		if errors.Is(err, ports.ErrResendTooSoon) {
			return meta, ErrCannotResendYet
		}
		return meta, ErrVerificationFailed
	}

	if err := deps.notifier.SendEmailVerification(ctx, ports.EmailVerificationMessage{
		UserID:     userID,
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		Code:       code,
		TTLSeconds: ttlSeconds,
		Locale:     normalizeLocale(locale),
	}); err != nil {
		return meta, ErrVerificationFailed
	}

	return meta, nil
}
