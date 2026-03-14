package usecase

import (
	"auth/internal/application/ports"
	"context"
	"errors"

	"github.com/google/uuid"
)

type LogoutUseCase struct {
	deps *dependencies
}

type LogoutAllUseCase struct {
	deps *dependencies
}

func (uc *LogoutUseCase) Execute(ctx context.Context, accessToken string) error {
	claims, err := uc.deps.tokens.ValidateToken(accessToken)
	if err != nil {
		return ErrUnauthorized
	}
	if err := uc.deps.tokens.EnsureTokenType(claims, ports.TokenTypeAccess); err != nil {
		return ErrUnauthorized
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return ErrUnauthorized
	}

	if err := uc.deps.sessions.Revoke(ctx, sessionID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return nil
}

func (uc *LogoutAllUseCase) Execute(ctx context.Context, accessToken string) error {
	claims, err := uc.deps.tokens.ValidateToken(accessToken)
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

	return uc.deps.sessions.RevokeAll(ctx, userID)
}
