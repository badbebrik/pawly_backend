package service

import (
	"auth/internal/domain/model"
	"auth/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) createUserWithProfile(ctx context.Context, user *model.User, locale, firstName, lastName string) error {
	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return repository.ErrConflict
		}
		return err
	}

	if err := s.profile.CreateProfile(ctx, user.ID, locale, firstName, lastName); err != nil {
		_ = s.users.Delete(ctx, user.ID)
		return ErrProfileCreationFailed
	}

	return nil
}

func (s *Service) compensateUserWithProfile(ctx context.Context, userID uuid.UUID) {
	if err := s.profile.DeleteProfile(ctx, userID); err != nil {
	}
	_ = s.users.Delete(ctx, userID)
}
