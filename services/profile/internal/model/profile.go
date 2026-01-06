package model

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserID        uuid.UUID      `json:"user_id"`
	FirstName     *string        `json:"first_name"`
	LastName      *string        `json:"last_name"`
	AvatarURL     *string        `json:"avatar_url"`
	Phone         *string        `json:"phone"`
	Locale        string         `json:"locale"`
	Timezone      string         `json:"timezone"`
	DateFormat    string         `json:"date_format"`
	Notifications map[string]any `json:"notifications"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
