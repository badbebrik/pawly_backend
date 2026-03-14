package ports

import (
	"context"

	"github.com/google/uuid"
)

type EmailVerificationMessage struct {
	UserID     uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	Code       string
	TTLSeconds int
	Locale     string
}

type PasswordResetMessage struct {
	UserID     uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	Code       string
	TTLSeconds int
	Locale     string
}

type WelcomeEmailMessage struct {
	UserID    uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Locale    string
}

type Notifier interface {
	SendEmailVerification(ctx context.Context, msg EmailVerificationMessage) error
	SendPasswordReset(ctx context.Context, msg PasswordResetMessage) error
	SendWelcomeEmail(ctx context.Context, msg WelcomeEmailMessage) error
}
