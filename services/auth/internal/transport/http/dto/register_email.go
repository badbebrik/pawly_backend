package dto

import "github.com/google/uuid"

type RegisterEmailRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Locale    string `json:"locale"`
	Timezone  string `json:"time_zone"`
}

type RegisterEmailResponse struct {
	UserID       uuid.UUID            `json:"user_id"`
	Verification VerificationResponse `json:"verification"`
}
