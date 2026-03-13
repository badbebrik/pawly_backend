package model

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserID       uuid.UUID  `json:"user_id"`
	FirstName    *string    `json:"first_name"`
	LastName     *string    `json:"last_name"`
	AvatarFileID *uuid.UUID `json:"avatar_file_id"`
	Locale       string     `json:"locale"`
	Timezone     string     `json:"time_zone"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
