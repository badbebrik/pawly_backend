package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/security"
	"context"
	"errors"
)

type PasswordResetRequestUseCase struct {
	deps *dependencies
}

type PasswordResetRequestInput struct {
	Email  string
	Locale string
}

func (uc *PasswordResetRequestUseCase) Execute(ctx context.Context, in PasswordResetRequestInput) error {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) {
		return ErrIncorrectFormat
	}

	user, err := uc.deps.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil
		}
		return err
	}
	if !user.CanRequestPasswordReset() {
		return nil
	}

	code, ttlSeconds, _, err := uc.deps.verification.RequestCode(ctx, email, "password_reset")
	if err != nil {
		if errors.Is(err, ports.ErrResendTooSoon) {
			return ErrCannotResendYet
		}
		return ErrVerificationFailed
	}

	if err := uc.deps.notifier.SendPasswordReset(ctx, ports.PasswordResetMessage{
		UserID:     user.ID,
		Email:      email,
		FirstName:  "",
		LastName:   "",
		Code:       code,
		TTLSeconds: ttlSeconds,
		Locale:     normalizeLocale(in.Locale),
	}); err != nil {
		return ErrVerificationFailed
	}

	return nil
}
