package dto

import "github.com/google/uuid"

type RegisterEmailRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Locale    string
}

type RegisterEmailResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Verification struct {
		Channel            string `json:"channel"`
		CodeTTLSeconds     int    `json:"code_ttl_seconds"`
		CanResendInSeconds int    `json:"can_resend_in_seconds"`
	} `json:"verification"`
}
