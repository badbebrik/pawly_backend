package ports

import (
	"context"

	"push/internal/domain/model"

	"github.com/google/uuid"
)

type UpsertDeviceTokenParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	Platform  string
	PushToken string
}

type DeleteDeviceTokenParams struct {
	UserID   uuid.UUID
	DeviceID string
}

type UpsertPetPushSettingsParams struct {
	UserID                uuid.UUID
	PetID                 uuid.UUID
	ScheduledItemsEnabled bool
}

type PushRepository interface {
	UpsertDeviceToken(ctx context.Context, in UpsertDeviceTokenParams) (*model.DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, in DeleteDeviceTokenParams) error
	ListDeviceTokensByUser(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error)
	GetPetPushSettings(ctx context.Context, userID, petID uuid.UUID) (*model.PetPushSettings, error)
	UpsertPetPushSettings(ctx context.Context, in UpsertPetPushSettingsParams) (*model.PetPushSettings, error)
}
