package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/security"
	"context"
	"errors"
	"regexp"
)

var resetCodeRe = regexp.MustCompile(`^\d{6}$`)

type PasswordResetVerify struct {
	deps *dependencies
}

type PasswordResetVerifyParams struct {
	Email string
	Code  string
}

type PasswordResetVerifyResult struct {
	ResetToken string
}

func (uc *PasswordResetVerify) Execute(ctx context.Context, in PasswordResetVerifyParams) (*PasswordResetVerifyResult, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) || !resetCodeRe.MatchString(in.Code) {
		return nil, ErrIncorrectFormat
	}

	user, err := uc.deps.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrVerificationCodeInvalid
		}
		return nil, err
	}
	if err := user.RequireActive(); err != nil {
		return nil, ErrUserBlocked
	}

	if err := uc.deps.verification.VerifyCode(ctx, email, "password_reset", in.Code); err != nil {
		switch {
		case errors.Is(err, ports.ErrCodeInvalid), errors.Is(err, ports.ErrCodeNotFound):
			return nil, ErrVerificationCodeInvalid
		case errors.Is(err, ports.ErrCodeExpired):
			return nil, ErrVerificationCodeExpired
		case errors.Is(err, ports.ErrTooManyAttempts):
			return nil, ErrVerificationTooMany
		default:
			return nil, ErrVerificationFailed
		}
	}

	resetToken, err := uc.deps.tokens.GeneratePasswordResetToken(user.ID.String(), email)
	if err != nil {
		return nil, err
	}

	return &PasswordResetVerifyResult{ResetToken: resetToken}, nil
}
