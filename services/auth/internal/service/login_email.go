package service

import (
	"auth/internal/domain/model"
	"auth/internal/infrastructure/rabbit"
	"auth/internal/repository"
	"auth/internal/security"
	"auth/internal/transport/http/middleware"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

type LoginEmailInput struct {
	Email    string
	Password string
}

type LoginEmailOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}

func (s *Service) LoginEmail(ctx context.Context, in LoginEmailInput) (*LoginEmailOutput, error) {
	email := security.NormalizeEmail(in.Email)

	if !security.ValidateEmail(email) || in.Password == "" {
		return nil, ErrIncorrectFormat
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidEmailOrPassword
		}
		return nil, err
	}

	if !u.IsActive {
		return nil, ErrUserBlocked
	}

	if !u.IsVerified {
		loc := middleware.LocaleFromCtx(ctx, "ru")

		code, ttlSeconds, _, verr := s.verification.RequestCode(ctx, email, "registration")
		if verr == nil {
			_ = s.notifier.SendEmailVerification(ctx, rabbit.EmailVerificationPayload{
				UserID:     u.ID,
				Email:      email,
				FirstName:  "",
				LastName:   "",
				Code:       code,
				TTLSeconds: ttlSeconds,
				Locale:     loc,
			})
		}
		return nil, ErrEmailNotVerified
	}

	if !u.HasPassword() {
		return nil, ErrInvalidEmailOrPassword
	}

	if err := security.ComparePassword(*u.PasswordHash, in.Password); err != nil {
		return nil, ErrInvalidEmailOrPassword
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

	return &LoginEmailOutput{
		UserID:       u.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
