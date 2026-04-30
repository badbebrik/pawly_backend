package dto

import "github.com/google/uuid"

type ProfileResponse struct {
	UserID            uuid.UUID `json:"user_id"`
	FirstName         *string   `json:"first_name"`
	LastName          *string   `json:"last_name"`
	AvatarDownloadURL *string   `json:"avatar_download_url"`
	Locale            string    `json:"locale"`
	Timezone          string    `json:"time_zone"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

type UpdatePreferencesRequest struct {
	Locale   *string `json:"locale"`
	Timezone *string `json:"time_zone"`
}

type BatchProfilesBriefRequest struct {
	UserIDs []string `json:"user_ids"`
}

type ProfileBriefResponse struct {
	UserID            uuid.UUID `json:"user_id"`
	FirstName         *string   `json:"first_name"`
	LastName          *string   `json:"last_name"`
	DisplayName       *string   `json:"display_name"`
	AvatarDownloadURL *string   `json:"avatar_download_url"`
}

type BatchProfilesBriefResponse struct {
	Items           []ProfileBriefResponse `json:"items"`
	NotFoundUserIDs []uuid.UUID            `json:"not_found_user_ids"`
}
