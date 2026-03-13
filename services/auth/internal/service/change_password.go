package service

import (
	"auth/internal/infrastructure/tokens"
	"auth/internal/repository"
	"auth/internal/security"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ChangePasswordInput struct {
	AccessToken string
	OldPassword string
	NewPassword string
}

func (s *Service) ChangePassword(ctx context.Context, in ChangePasswordInput) error {
	if in.AccessToken == "" || in.OldPassword == "" {
		return ErrIncorrectFormat
	}
	if !security.ValidatePassword(in.NewPassword) {
		return ErrWeakPassword
	}

	p, err := s.jwt.ValidateToken(in.AccessToken)
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

	u, err := s.users.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if !u.IsActive {
		return ErrUserBlocked
	}
	if !u.HasPassword() {
		return ErrInvalidEmailOrPassword
	}
	if err := security.ComparePassword(*u.PasswordHash, in.OldPassword); err != nil {
		return ErrInvalidEmailOrPassword
	}

	hash, err := security.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePasswordHash(ctx, uid, hash); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := s.sessions.RevokeAll(ctx, uid); err != nil {
		return err
	}

	return nil
}
