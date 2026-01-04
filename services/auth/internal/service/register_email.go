package service

import (
	"github.com/google/uuid"
)

type RegisterEmailInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Locale    string
}

type RegisterEmailOutput struct {
	UserID       uuid.UUID
	Verification struct {
		Channel            string
		CodeTTLSeconds     int
		CanResendInSeconds int
	}
}
