package service

import (
	"context"
	"strings"

	"push/internal/model"
	"push/internal/repository"

	"github.com/google/uuid"
)

type Service struct {
	repo repository.PushRepository
}

func New(repo repository.PushRepository) *Service {
	return &Service{repo: repo}
}

type RegisterDeviceTokenParams struct {
	UserID    uuid.UUID
	DeviceID  string
	Platform  string
	PushToken string
}

type DeleteDeviceTokenParams struct {
	UserID   uuid.UUID
	DeviceID string
}

type GetPetPushSettingsParams struct {
	UserID uuid.UUID
	PetID  uuid.UUID
}

type UpdatePetPushSettingsParams struct {
	UserID                uuid.UUID
	PetID                 uuid.UUID
	ScheduledItemsEnabled bool
}

type ListEligibleDeviceTokensParams struct {
	UserID uuid.UUID
	PetID  uuid.UUID
}

func (s *Service) RegisterDeviceToken(ctx context.Context, in RegisterDeviceTokenParams) (*model.DeviceToken, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(in.DeviceID) == "" || strings.TrimSpace(in.PushToken) == "" {
		return nil, ErrValidation
	}
	if in.Platform != model.PlatformIOS && in.Platform != model.PlatformAndroid {
		return nil, ErrValidation
	}

	item, err := s.repo.UpsertDeviceToken(ctx, repository.UpsertDeviceTokenInput{
		ID:        uuid.New(),
		UserID:    in.UserID,
		DeviceID:  strings.TrimSpace(in.DeviceID),
		Platform:  in.Platform,
		PushToken: strings.TrimSpace(in.PushToken),
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	return item, nil
}

func (s *Service) DeleteDeviceToken(ctx context.Context, in DeleteDeviceTokenParams) error {
	if in.UserID == uuid.Nil {
		return ErrForbidden
	}
	if strings.TrimSpace(in.DeviceID) == "" {
		return ErrValidation
	}
	return mapRepoError(s.repo.DeleteDeviceToken(ctx, repository.DeleteDeviceTokenInput{
		UserID:   in.UserID,
		DeviceID: strings.TrimSpace(in.DeviceID),
	}))
}

func (s *Service) GetPetPushSettings(ctx context.Context, in GetPetPushSettingsParams) (*model.PetPushSettings, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrForbidden
	}
	if in.PetID == uuid.Nil {
		return nil, ErrValidation
	}

	item, err := s.repo.GetPetPushSettings(ctx, in.UserID, in.PetID)
	if err != nil {
		if err == repository.ErrNotFound {
			return &model.PetPushSettings{
				UserID:                in.UserID,
				PetID:                 in.PetID,
				ScheduledItemsEnabled: true,
			}, nil
		}
		return nil, mapRepoError(err)
	}
	return item, nil
}

func (s *Service) UpdatePetPushSettings(ctx context.Context, in UpdatePetPushSettingsParams) (*model.PetPushSettings, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrForbidden
	}
	if in.PetID == uuid.Nil {
		return nil, ErrValidation
	}

	item, err := s.repo.UpsertPetPushSettings(ctx, repository.UpsertPetPushSettingsInput{
		UserID:                in.UserID,
		PetID:                 in.PetID,
		ScheduledItemsEnabled: in.ScheduledItemsEnabled,
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	return item, nil
}

func (s *Service) ListEligibleDeviceTokens(ctx context.Context, in ListEligibleDeviceTokensParams) ([]model.DeviceToken, error) {
	if in.UserID == uuid.Nil || in.PetID == uuid.Nil {
		return nil, ErrValidation
	}

	settings, err := s.GetPetPushSettings(ctx, GetPetPushSettingsParams{
		UserID: in.UserID,
		PetID:  in.PetID,
	})
	if err != nil {
		return nil, err
	}
	if !settings.ScheduledItemsEnabled {
		return []model.DeviceToken{}, nil
	}

	items, err := s.repo.ListDeviceTokensByUser(ctx, in.UserID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	return items, nil
}

func mapRepoError(err error) error {
	switch err {
	case nil:
		return nil
	case repository.ErrNotFound:
		return ErrNotFound
	case repository.ErrConflict:
		return ErrConflict
	default:
		return err
	}
}
