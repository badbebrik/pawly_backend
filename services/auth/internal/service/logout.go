package service

import (
	"auth/internal/infrastructure/tokens"
	"auth/internal/repository"
	"context"
	"errors"
	"github.com/google/uuid"
)

func (s *Service) Logout(ctx context.Context, accessToken string) error {
	p, err := s.jwt.ValidateToken(accessToken)

	if err != nil {
		return ErrUnauthorized
	}

	if err := s.jwt.EnsureTokenType(p, tokens.TokenTypeAccess); err != nil {
		return ErrUnauthorized
	}

	sid, err := uuid.Parse(p.SessionID)
	if err != nil {
		return ErrUnauthorized
	}

	if err := s.sessions.Revoke(ctx, sid); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, accessToken string) error {
	p, err := s.jwt.ValidateToken(accessToken)
	if err != nil {
		return ErrUnauthorized
	}
	if err := s.jwt.EnsureTokenType(p, tokens.TokenTypeAccess); err != nil {
		return ErrUnauthorized
	}

	uid, err := uuid.Parse(p.Sub)
	if err != nil {
		return ErrUnauthorized
	}

	return s.sessions.RevokeAll(ctx, uid)
}
