package usecase

import (
	"context"
	"strings"

	"push/internal/application/ports"
	"push/internal/domain/model"

	"github.com/google/uuid"
)

type Set struct {
	repo ports.PushRepository
}

func New(repo ports.PushRepository) *Set {
	return &Set{repo: repo}
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

func (s *Set) RegisterDeviceToken(ctx context.Context, in RegisterDeviceTokenParams) (*model.DeviceToken, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(in.DeviceID) == "" || strings.TrimSpace(in.PushToken) == "" {
		return nil, ErrInvalidInput
	}
	platform := strings.ToUpper(strings.TrimSpace(in.Platform))
	if platform != model.PlatformIOS && platform != model.PlatformAndroid {
		return nil, ErrInvalidInput
	}

	item, err := s.repo.UpsertDeviceToken(ctx, ports.UpsertDeviceTokenParams{
		ID:        uuid.New(),
		UserID:    in.UserID,
		DeviceID:  strings.TrimSpace(in.DeviceID),
		Platform:  platform,
		PushToken: strings.TrimSpace(in.PushToken),
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	return item, nil
}

func (s *Set) DeleteDeviceToken(ctx context.Context, in DeleteDeviceTokenParams) error {
	if in.UserID == uuid.Nil {
		return ErrForbidden
	}
	if strings.TrimSpace(in.DeviceID) == "" {
		return ErrInvalidInput
	}
	return mapRepoError(s.repo.DeleteDeviceToken(ctx, ports.DeleteDeviceTokenParams{
		UserID:   in.UserID,
		DeviceID: strings.TrimSpace(in.DeviceID),
	}))
}

func (s *Set) GetPetPushSettings(ctx context.Context, in GetPetPushSettingsParams) (*model.PetPushSettings, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrForbidden
	}
	if in.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	item, err := s.repo.GetPetPushSettings(ctx, in.UserID, in.PetID)
	if err != nil {
		if err == ports.ErrNotFound {
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

func (s *Set) UpdatePetPushSettings(ctx context.Context, in UpdatePetPushSettingsParams) (*model.PetPushSettings, error) {
	if in.UserID == uuid.Nil {
		return nil, ErrForbidden
	}
	if in.PetID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	item, err := s.repo.UpsertPetPushSettings(ctx, ports.UpsertPetPushSettingsParams{
		UserID:                in.UserID,
		PetID:                 in.PetID,
		ScheduledItemsEnabled: in.ScheduledItemsEnabled,
	})
	if err != nil {
		return nil, mapRepoError(err)
	}
	return item, nil
}

func (s *Set) ListEligibleDeviceTokens(ctx context.Context, in ListEligibleDeviceTokensParams) ([]model.DeviceToken, error) {
	if in.UserID == uuid.Nil || in.PetID == uuid.Nil {
		return nil, ErrInvalidInput
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
	case ports.ErrNotFound:
		return ErrNotFound
	case ports.ErrConflict:
		return ErrConflict
	default:
		return err
	}
}
