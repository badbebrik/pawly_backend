package usecase

import (
	"auth/internal/application/ports"
	"auth/internal/domain/model"
	"auth/internal/security"
	"context"
	"errors"

	"github.com/google/uuid"
)

type RefreshUseCase struct {
	deps *dependencies
}

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (uc *RefreshUseCase) Execute(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	if in.RefreshToken == "" {
		return nil, ErrIncorrectFormat
	}

	claims, err := uc.deps.tokens.ValidateToken(in.RefreshToken)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if err := uc.deps.tokens.EnsureTokenType(claims, ports.TokenTypeRefresh); err != nil {
		return nil, ErrUnauthorized
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrUnauthorized
	}

	session, err := uc.deps.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if session.UserID != userID {
		return nil, ErrUnauthorized
	}
	inHash := security.HashTokenSHA256(in.RefreshToken)
	if err := session.CanRefresh(uc.deps.now(), inHash); err != nil {
		switch {
		case errors.Is(err, model.ErrSessionRevoked):
			return nil, ErrSessionRevoked
		case errors.Is(err, model.ErrSessionExpired):
			return nil, ErrSessionExpired
		case errors.Is(err, model.ErrRefreshTokenMismatch):
			return nil, ErrRefreshMismatch
		default:
			return nil, err
		}
	}

	pair, err := uc.deps.rotateSession(ctx, userID, sessionID, inHash)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrRefreshMismatch
		}
		return nil, err
	}

	return &RefreshOutput{
		UserID:       userID,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    int(uc.deps.tokens.AccessTTL().Seconds()),
	}, nil
}
