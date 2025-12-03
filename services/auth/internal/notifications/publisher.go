package notifications

import (
	"context"
	"github.com/google/uuid"
)

type Publisher interface {
	SendEmailVerification(ctx context.Context, p EmailVerificationPayload) error
	SendPasswordReset(ctx context.Context, p PasswordResetPayload) error
	SendWelcomeEmail(ctx context.Context, p WelcomeEmailPayload) error
}

type NotificationEvent struct {
	Event    string                 `json:"event"`
	UserID   uuid.UUID              `json:"user_id"`
	Locale   string                 `json:"locale"`
	Channels []string               `json:"channels"`
	Data     map[string]interface{} `json:"data"`
}

func NewEvent(event string, userID uuid.UUID, locale string, channels []string, data map[string]interface{}) NotificationEvent {
	return NotificationEvent{
		Event:    event,
		UserID:   userID,
		Locale:   locale,
		Channels: channels,
		Data:     data,
	}
}

type EmailVerificationPayload struct {
	UserID     uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	Code       string
	TTLSeconds int
	Locale     string
}

type PasswordResetPayload struct {
	UserID     uuid.UUID
	Email      string
	FirstName  string
	LastName   string
	Code       string
	TTLSeconds int
	Locale     string
}

type WelcomeEmailPayload struct {
	UserID    uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Locale    string
}

func (p *RabbitPublisher) SendEmailVerification(ctx context.Context, a EmailVerificationPayload) error {
	ev := NewEvent(
		"EMAIL_VERIFICATION_REQUESTED",
		a.UserID,
		a.Locale,
		[]string{"email"},
		map[string]interface{}{
			"email":       a.Email,
			"first_name":  a.FirstName,
			"last_name":   a.LastName,
			"code":        a.Code,
			"ttl_seconds": a.TTLSeconds,
			"purpose":     "registration",
		},
	)

	return p.publish(ctx, ev)
}

func (p *RabbitPublisher) SendPasswordReset(ctx context.Context, a PasswordResetPayload) error {
	ev := NewEvent(
		"PASSWORD_RESET_REQUESTED",
		a.UserID,
		a.Locale,
		[]string{"email"},
		map[string]interface{}{
			"email":       a.Email,
			"first_name":  a.FirstName,
			"last_name":   a.LastName,
			"code":        a.Code,
			"ttl_seconds": a.TTLSeconds,
		},
	)

	return p.publish(ctx, ev)
}

func (p *RabbitPublisher) SendWelcomeEmail(ctx context.Context, a WelcomeEmailPayload) error {
	ev := NewEvent(
		"WELCOME_EMAIL",
		a.UserID,
		a.Locale,
		[]string{"email"},
		map[string]interface{}{
			"email":      a.Email,
			"first_name": a.FirstName,
			"last_name":  a.LastName,
		},
	)

	return p.publish(ctx, ev)
}
