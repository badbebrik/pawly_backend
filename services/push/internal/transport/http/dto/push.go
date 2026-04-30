package dto

import (
	"time"

	"github.com/google/uuid"
)

type UpsertDeviceTokenRequest struct {
	DeviceID  string `json:"device_id"`
	Platform  string `json:"platform"`
	PushToken string `json:"push_token"`
}

type PetPushSettingsRequest struct {
	ScheduledItemsEnabled bool `json:"scheduled_items_enabled"`
}

type DeviceTokenResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	Platform  string    `json:"platform"`
	PushToken string    `json:"push_token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeviceTokenEnvelopeResponse struct {
	Item DeviceTokenResponse `json:"item"`
}

type PetPushSettingsResponse struct {
	PetID                 uuid.UUID `json:"pet_id"`
	ScheduledItemsEnabled bool      `json:"scheduled_items_enabled"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type PetPushSettingsEnvelopeResponse struct {
	Item PetPushSettingsResponse `json:"item"`
}
