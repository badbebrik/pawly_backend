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
	"regexp"
	"time"
)

var verificationCodeRe = regexp.MustCompile(`^\d{6}$`)

type VerifyEmailInput struct {
	Email string
	Code  string
}

type VerifyEmailOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (s *Service) VerifyEmail(ctx context.Context, in VerifyEmailInput) (*VerifyEmailOutput, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) || !verificationCodeRe.MatchString(in.Code) {
		return nil, ErrIncorrectFormat
	}

	loc := middleware.LocaleFromCtx(ctx, "ru")

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

	if !u.IsVerified {
		if err := s.verification.VerifyCode(ctx, email, "registration", in.Code); err != nil {
			switch {
			case errors.Is(err, verification.ErrCodeInvalid),
				errors.Is(err, verification.ErrCodeNotFound):
				return nil, ErrVerificationCodeInvalid
			case errors.Is(err, verification.ErrCodeExpired):
				return nil, ErrVerificationCodeExpired
			case errors.Is(err, verification.ErrTooManyAttempts):
				return nil, ErrVerificationTooMany
			default:
				return nil, ErrVerificationFailed
			}
		}

		if err := s.users.SetVerified(ctx, u.ID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}

		_ = s.notifier.SendWelcomeEmail(ctx, rabbit.WelcomeEmailPayload{
			UserID:    u.ID,
			Email:     u.Email,
			FirstName: "",
			LastName:  "",
			Locale:    loc,
		})
	}

	sessionID := uuid.New()
	refreshToken, err := s.jwt.GenerateRefreshToken(u.ID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}
	accessToken, err := s.jwt.GenerateAccessToken(u.ID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}

	hash := security.HashTokenSHA256(refreshToken)
	expiresAt := time.Now().Add(time.Duration(s.refreshTTLDays) * 24 * time.Hour)

	sess := &model.Session{
		ID:               sessionID,
		UserID:           u.ID,
		RefreshTokenHash: hash,
		ExpiresAt:        expiresAt,
		IsRevoked:        false,
	}

	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, err
	}

	_ = s.users.UpdateLastLoginAt(ctx, u.ID, time.Now())

	return &VerifyEmailOutput{
		UserID:       u.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.accessTTLSeconds,
	}, nil
}
