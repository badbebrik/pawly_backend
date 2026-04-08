package repository

import (
	"context"

	"push/internal/model"

	"github.com/google/uuid"
)

type UpsertDeviceTokenInput struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	Platform  string
	PushToken string
}

type DeleteDeviceTokenInput struct {
	UserID   uuid.UUID
	DeviceID string
}

type UpsertPetPushSettingsInput struct {
	UserID                uuid.UUID
	PetID                 uuid.UUID
	ScheduledItemsEnabled bool
}

type PushRepository interface {
	UpsertDeviceToken(ctx context.Context, in UpsertDeviceTokenInput) (*model.DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, in DeleteDeviceTokenInput) error
	ListDeviceTokensByUser(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error)
	GetPetPushSettings(ctx context.Context, userID, petID uuid.UUID) (*model.PetPushSettings, error)
	UpsertPetPushSettings(ctx context.Context, in UpsertPetPushSettingsInput) (*model.PetPushSettings, error)
}
