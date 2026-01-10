package service

import "github.com/google/uuid"

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}
