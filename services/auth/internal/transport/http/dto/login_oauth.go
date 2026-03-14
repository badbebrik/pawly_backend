package dto

import "github.com/google/uuid"

type LoginOAuthRequest struct {
	Provider string `json:"provider"`
	IDToken  string `json:"id_token"`
	Locale   string `json:"locale"`
	Timezone string `json:"time_zone"`
}

type LoginOAuthResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
}
