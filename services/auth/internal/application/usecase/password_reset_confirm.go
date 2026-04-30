package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/security"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PasswordResetConfirm struct {
	deps *dependencies
}

type PasswordResetConfirmParams struct {
	ResetToken  string
	NewPassword string
}

func (uc *PasswordResetConfirm) Execute(ctx context.Context, in PasswordResetConfirmParams) error {
	if in.ResetToken == "" || !security.ValidatePassword(in.NewPassword) {
		return ErrIncorrectFormat
	}

	claims, err := uc.deps.tokens.ValidateToken(in.ResetToken)
	if err != nil {
		return ErrUnauthorized
	}
	if err := uc.deps.tokens.EnsureTokenType(claims, ports.TokenTypeReset); err != nil {
		return ErrUnauthorized
	}
	if claims.SessionID == "" {
		return ErrUnauthorized
	}

	ttl := time.Until(claims.ExpiresAt)
	consumed, err := uc.deps.resetTokens.ConsumeOnce(ctx, claims.SessionID, ttl)
	if err != nil {
		return err
	}
	if !consumed {
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
	if err := user.RequireActive(); err != nil {
		return ErrUserBlocked
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
