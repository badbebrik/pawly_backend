package repository

import (
	"auth/internal/domain/model"
	"context"
	"github.com/google/uuid"
)

type DeviceRepository interface {
	Upsert(ctx context.Context, d *model.Device) error
	GetByUserAndDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*model.Device, error)
	SetActive(ctx context.Context, userID uuid.UUID, deviceID string, active bool) error
	DeactivateByFCMToken(ctx context.Context, token string) error
	DeactivateAllByUserID(ctx context.Context, userID uuid.UUID) error
}
