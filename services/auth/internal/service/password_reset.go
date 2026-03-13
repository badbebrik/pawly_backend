package service

import (
	"auth/internal/infrastructure/rabbit"
	"auth/internal/infrastructure/tokens"
	"auth/internal/repository"
	"auth/internal/security"
	"auth/internal/verification"
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
	"time"
)

var resetCodeRe = regexp.MustCompile(`^\d{6}$`)

type PasswordResetRequestInput struct {
	Email string
}

type PasswordResetVerifyInput struct {
	Email string
	Code  string
}

type PasswordResetVerifyOutput struct {
	ResetToken string
}

type PasswordResetConfirmInput struct {
	ResetToken  string
	NewPassword string
}

func (s *Service) RequestPasswordReset(ctx context.Context, in PasswordResetRequestInput) error {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) {
		return ErrIncorrectFormat
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}
	if !u.IsVerified || !u.IsActive {
		return nil
	}

	code, ttlSeconds, _, err := s.verification.RequestCode(ctx, email, "password_reset")
	if err != nil {
		if errors.Is(err, verification.ErrResendTooSoon) {
			return ErrCannotResendYet
		}
		return ErrVerificationFailed
	}

	if err := s.notifier.SendPasswordReset(ctx, rabbit.PasswordResetPayload{
		UserID:     u.ID,
		Email:      email,
		FirstName:  "",
		LastName:   "",
		Code:       code,
		TTLSeconds: ttlSeconds,
		Locale:     "ru",
	}); err != nil {
		return ErrVerificationFailed
	}

	return nil
}

func (s *Service) VerifyPasswordResetCode(ctx context.Context, in PasswordResetVerifyInput) (*PasswordResetVerifyOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) || !resetCodeRe.MatchString(in.Code) {
		return nil, ErrIncorrectFormat
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrVerificationCodeInvalid
		}
		return nil, err
	}
	if !u.IsActive {
		return nil, ErrUserBlocked
	}

	if err := s.verification.VerifyCode(ctx, email, "password_reset", in.Code); err != nil {
		switch {
		case errors.Is(err, verification.ErrCodeInvalid), errors.Is(err, verification.ErrCodeNotFound):
			return nil, ErrVerificationCodeInvalid
		case errors.Is(err, verification.ErrCodeExpired):
			return nil, ErrVerificationCodeExpired
		case errors.Is(err, verification.ErrTooManyAttempts):
			return nil, ErrVerificationTooMany
		default:
			return nil, ErrVerificationFailed
		}
	}

	resetToken, err := s.jwt.GeneratePasswordResetToken(u.ID.String(), email)
	if err != nil {
		return nil, err
	}

	return &PasswordResetVerifyOutput{ResetToken: resetToken}, nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, in PasswordResetConfirmInput) error {
	if in.ResetToken == "" || !security.ValidatePassword(in.NewPassword) {
		return ErrIncorrectFormat
	}

	p, err := s.jwt.ValidateToken(in.ResetToken)
	if err != nil {
		return ErrUnauthorized
	}
	if err := s.jwt.EnsureTokenType(p, tokens.TokenTypeReset); err != nil {
		return ErrUnauthorized
	}
	if p.SessionID == "" {
		return ErrUnauthorized
	}

	ttl := time.Until(time.Unix(p.Exp, 0))
	consumed, err := s.resetTokens.ConsumeOnce(ctx, p.SessionID, ttl)
	if err != nil {
		return err
	}
	if !consumed {
		return ErrUnauthorized
	}

	uid, err := uuid.Parse(p.Sub)
	if err != nil {
		return ErrUnauthorized
	}

	u, err := s.users.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if !u.IsActive {
		return ErrUserBlocked
	}

	hash, err := security.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePasswordHash(ctx, uid, hash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.sessions.RevokeAll(ctx, uid); err != nil {
		return err
	}

	return nil
}
