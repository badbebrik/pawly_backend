package model

import (
	"github.com/google/uuid"
	"time"
)

type Profile struct {
	UserID        uuid.UUID             `json:"user_id"`
	FirstName     *string               `json:"first_name"`
	LastName      *string               `json:"last_name"`
	Phone         *string               `json:"phone"`
	AvatarFileID  *uuid.UUID            `json:"avatar_file_id"`
	Locale        string                `json:"locale"`
	Timezone      string                `json:"time_zone"`
	DateFormat    string                `json:"date_format"`
	PublicContact PublicContactSettings `json:"public_contact_settings"`
	ExtraContacts ExtraContacts         `json:"extra_contacts"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type PublicContactSettings struct {
	ShowOwnerName             bool    `json:"show_owner_name"`
	ShowPhone                 bool    `json:"show_phone"`
	ShowEmail                 bool    `json:"show_email"`
	ShowExtraContacts         bool    `json:"show_extra_contacts"`
	PublicDisplayNameOverride *string `json:"public_display_name_override"`
}

type ExtraContacts map[string]string
