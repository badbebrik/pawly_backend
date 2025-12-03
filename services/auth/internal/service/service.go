package auth

import (
	dto "auth/internal/dto/register_email"
	"auth/internal/model"
	"auth/internal/notifications"
	"auth/internal/repository/device_repo"
	"auth/internal/repository/oauth_identity_repo"
	"auth/internal/repository/session_repo"
	"auth/internal/repository/user_repo"
	"auth/internal/tokens"
	"auth/internal/util"
	"auth/internal/verification"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

type Service struct {
	users        user_repo.Repository
	sessions     session_repo.Repository
	oauth        oauth_identity_repo.Repository
	devices      device_repo.Repository
	verification verification.Repository
	notifier     notifications.Publisher
	jwt          *tokens.JWTService
	profiles     ProfileClient
}

type ProfileClient interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error
}

func NewService(
	users user_repo.Repository,
	sessions session_repo.Repository,
	oauth oauth_identity_repo.Repository,
	devices device_repo.Repository,
	verification verification.Repository,
	notifier notifications.Publisher,
	jwt *tokens.JWTService,
	profiles ProfileClient,
) *Service {
	return &Service{
		users:        users,
		sessions:     sessions,
		oauth:        oauth,
		devices:      devices,
		verification: verification,
		notifier:     notifier,
		jwt:          jwt,
		profiles:     profiles,
	}
}

func (s *Service) RegisterEmail(ctx context.Context, req dto.RegisterEmailRequest) (*dto.RegisterEmailResponse, error) {
	email := util.NormalizeEmail(req.Email)

	if !util.ValidateEmail(email) || !util.ValidatePassword(req.Password) {
		return nil, ErrIncorrectFormat
	}

	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyTaken
	} else if err != nil && !errors.Is(user_repo.ErrUserNotFound, err) {
		return nil, err
	}

	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: &passwordHash,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.users.CreateUserWithProfile(ctx, user, func() error {
		return s.profiles.CreateProfile(ctx, user.ID, req.FirstName, req.LastName)
	})
	if err != nil {
		return nil, err
	}

	code, ttlSec, resendSec, err := s.verification.RequestCode(ctx, email, "registration")
	if err != nil {
		_ = s.users.SoftDelete(ctx, user.ID)
		return nil, ErrVerificationFailed
	}

	if s.notifier != nil {
		_ = s.notifier.SendEmailVerification(ctx, notifications.EmailVerificationPayload{
			UserID:     user.ID,
			Email:      email,
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			Code:       code,
			TTLSeconds: ttlSec,
			Locale:     defaultLocale(req.Locale),
		})
	}

	resp := &dto.RegisterEmailResponse{
		UserID: user.ID,
	}
	resp.Verification.Channel = "email"
	resp.Verification.CodeTTLSeconds = ttlSec
	resp.Verification.CanResendInSeconds = resendSec

	return resp, nil
}

func defaultLocale(loc string) string {
	if loc == "" {
		return "ru"
	}
	return loc
}
