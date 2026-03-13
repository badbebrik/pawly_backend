package outbox

import (
	"auth/internal/infrastructure/rabbit"
	"auth/internal/repository"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type Publisher struct {
	repo repository.OutboxRepository
}

func NewPublisher(repo repository.OutboxRepository) *Publisher {
	return &Publisher{repo: repo}
}

func (p *Publisher) SendEmailVerification(ctx context.Context, a rabbit.EmailVerificationPayload) error {
	ev := rabbit.NewEvent(
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
	return p.enqueue(ctx, ev)
}

func (p *Publisher) SendPasswordReset(ctx context.Context, a rabbit.PasswordResetPayload) error {
	ev := rabbit.NewEvent(
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
	return p.enqueue(ctx, ev)
}

func (p *Publisher) SendWelcomeEmail(ctx context.Context, a rabbit.WelcomeEmailPayload) error {
	ev := rabbit.NewEvent(
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
	return p.enqueue(ctx, ev)
}

func (p *Publisher) enqueue(ctx context.Context, ev rabbit.NotificationEvent) error {
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	return p.repo.Create(ctx, repository.OutboxEvent{
		ID:        uuid.New(),
		EventType: ev.Event,
		Payload:   body,
		Status:    "PENDING",
	})
}
