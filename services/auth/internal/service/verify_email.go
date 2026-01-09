package service

import "github.com/google/uuid"

type VerifyEmailInput struct {
	Email  string
	Code   string
	Locale string
}

type VerifyEmailOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}
