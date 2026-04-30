package dto

import "github.com/google/uuid"

type VerificationResponse struct {
	Channel            string `json:"channel"`
	CodeTTLSeconds     int    `json:"code_ttl_seconds"`
	CanResendInSeconds int    `json:"can_resend_in_seconds"`
}

type SessionResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type EmptyResponse struct{}
