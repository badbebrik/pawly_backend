package service

import "github.com/google/uuid"

type LoginEmailInput struct {
	Email    string
	Password string
}

type LoginEmailOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}
