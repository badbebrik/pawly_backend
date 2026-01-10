package service

import (
	"auth/internal/infrastructure/tokens"
	"auth/internal/repository"
	"auth/internal/security"
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}

func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	if in.RefreshToken == "" {
		return nil, ErrIncorrectFormat
	}

	p, err := s.jwt.ValidateToken(in.RefreshToken)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if err := s.jwt.EnsureTokenType(p, tokens.TokenTypeRefresh); err != nil {
		return nil, ErrUnauthorized
	}

	sessionID, err := uuid.Parse(p.SessionID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	userID, err := uuid.Parse(p.Sub)
	if err != nil {
		return nil, ErrUnauthorized
	}

	sess, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if sess.UserID != userID {
		return nil, ErrUnauthorized
	}

	if sess.IsRevoked {
		return nil, ErrSessionRevoked
	}

	if time.Now().After(sess.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	inHash := security.HashTokenSHA256(in.RefreshToken)
	if inHash != sess.RefreshTokenHash {
		return nil, ErrRefreshMismatch
	}

	newRefresh, err := s.jwt.GenerateRefreshToken(userID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}
	newAccess, err := s.jwt.GenerateAccessToken(userID.String(), sessionID.String())
	if err != nil {
		return nil, err
	}

	newHash := security.HashTokenSHA256(newRefresh)
	newExpires := time.Now().Add(time.Duration(s.refreshTTLDays) * 24 * time.Hour)

	if err := s.sessions.UpdateRefreshToken(ctx, sessionID, newHash, newExpires); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	_ = s.users.UpdateLastLoginAt(ctx, userID, time.Now())

	return &RefreshOutput{
		UserID:       userID,
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}
