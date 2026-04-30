package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"

	"github.com/google/uuid"
)

type LoginEmail struct {
	deps *dependencies
}

type LoginEmailParams struct {
	Email    string
	Password string
	Locale   string
}

type LoginEmailResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (uc *LoginEmail) Execute(ctx context.Context, in LoginEmailParams) (*LoginEmailResult, error) {
	email := security.NormalizeEmail(in.Email)
	if !security.ValidateEmail(email) || in.Password == "" {
		return nil, ErrIncorrectFormat
	}

	user, err := uc.deps.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrInvalidEmailOrPassword
		}
		return nil, err
	}
	if err := user.RequireActive(); err != nil {
		return nil, ErrUserBlocked
	}
	if err := user.RequireVerified(); err != nil {
		_, _ = sendRegistrationVerification(ctx, uc.deps, user.ID, email, "", "", in.Locale)
		return nil, ErrEmailNotVerified
	}
	if err := user.RequirePasswordAccount(); err != nil {
		if errors.Is(err, model.ErrPasswordAuthUnavailable) {
			return nil, ErrInvalidEmailOrPassword
		}
		return nil, ErrInvalidEmailOrPassword
	}
	if err := security.ComparePassword(*user.PasswordHash, in.Password); err != nil {
		return nil, ErrInvalidEmailOrPassword
	}

	pair, err := uc.deps.createSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginEmailResult{
		UserID:       user.ID,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    int(uc.deps.tokens.AccessTTL().Seconds()),
	}, nil
}
