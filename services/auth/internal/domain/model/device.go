package model

import (
	"github.com/google/uuid"
	"time"
)

type Device struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	Platform   byte      `json:"platform"`
	AppVersion string    `json:"app_version"`
	Locale     string    `json:"locale"`
	FCMToken   *string   `json:"fcm_token"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

const (
	PlatformIOS     byte = 1
	PlatformAndroid byte = 2
	PlatformWeb     byte = 3
)
