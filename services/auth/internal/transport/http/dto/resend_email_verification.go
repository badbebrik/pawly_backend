package dto

import "github.com/google/uuid"

type ResendEmailVerificationRequest struct {
	Email string `json:"email"`
}

type ResendEmailVerificationResponse struct {
	UserID       uuid.UUID            `json:"user_id"`
	Verification VerificationResponse `json:"verification"`
}
