package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ChangePasswordUseCase struct {
	deps *dependencies
}

type ChangePasswordInput struct {
	AccessToken string
	OldPassword string
	NewPassword string
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, in ChangePasswordInput) error {
	if in.AccessToken == "" || in.OldPassword == "" {
		return ErrIncorrectFormat
	}
	if !security.ValidatePassword(in.NewPassword) {
		return ErrWeakPassword
	}

	claims, err := uc.deps.tokens.ValidateToken(in.AccessToken)
	if err != nil {
		return ErrUnauthorized
	}
	if err := uc.deps.tokens.EnsureTokenType(claims, ports.TokenTypeAccess); err != nil {
		return ErrUnauthorized
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return ErrUnauthorized
	}

	user, err := uc.deps.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := user.CanChangePassword(); err != nil {
		switch {
		case errors.Is(err, model.ErrUserInactive):
			return ErrUserBlocked
		case errors.Is(err, model.ErrPasswordAuthUnavailable):
			return ErrInvalidEmailOrPassword
		default:
			return err
		}
	}
	if err := security.ComparePassword(*user.PasswordHash, in.OldPassword); err != nil {
		return ErrInvalidEmailOrPassword
	}

	hash, err := security.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	if err := uc.deps.users.UpdatePasswordHash(ctx, userID, hash); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return uc.deps.sessions.RevokeAll(ctx, userID)
}
