package dto

import "github.com/google/uuid"

type LoginEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginEmailResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
}
