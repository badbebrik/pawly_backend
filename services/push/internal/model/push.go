package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	PlatformIOS     = "IOS"
	PlatformAndroid = "ANDROID"
)

type DeviceToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DeviceID  string
	Platform  string
	PushToken string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PetPushSettings struct {
	UserID                uuid.UUID
	PetID                 uuid.UUID
	ScheduledItemsEnabled bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
