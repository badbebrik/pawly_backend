package dto

import "github.com/google/uuid"

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyEmailResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
}
