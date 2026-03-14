package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/security"
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
)

var verificationCodeRe = regexp.MustCompile(`^\d{6}$`)

type VerifyEmailUseCase struct {
	deps *dependencies
}

type VerifyEmailInput struct {
	Email  string
	Code   string
	Locale string
}

type VerifyEmailOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, in VerifyEmailInput) (*VerifyEmailOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) || !verificationCodeRe.MatchString(in.Code) {
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

	if !user.IsVerified {
		if err := uc.deps.verification.VerifyCode(ctx, email, "registration", in.Code); err != nil {
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

		if err := uc.deps.users.SetVerified(ctx, user.ID); err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		user.MarkVerified()

		_ = uc.deps.notifier.SendWelcomeEmail(ctx, ports.WelcomeEmailMessage{
			UserID:    user.ID,
			Email:     user.Email,
			FirstName: "",
			LastName:  "",
			Locale:    normalizeLocale(in.Locale),
		})
	}

	pair, err := uc.deps.createSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &VerifyEmailOutput{
		UserID:       user.ID,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    uc.deps.accessTTLSeconds,
	}, nil
}
