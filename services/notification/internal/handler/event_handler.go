package handler

import (
	"context"
	"encoding/json"

	"notification/internal/model"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type EmailPublisher interface {
	Publish(ctx context.Context, job model.EmailJob) error
}

type EventHandler struct {
	email EmailPublisher
}

func NewEventHandler(email EmailPublisher) *EventHandler {
	return &EventHandler{email: email}
}

func (h *EventHandler) Handle(ctx context.Context, msg amqp091.Delivery) {
	var ev model.NotificationEvent
	if err := json.Unmarshal(msg.Body, &ev); err != nil {
		log.Warn().Err(err).Msg("invalid event json")
		_ = msg.Ack(false)
		return
	}

	switch ev.Event {
	case "EMAIL_VERIFICATION_REQUESTED":
		h.handleEmailVerification(ctx, ev, msg)
	case "PASSWORD_RESET_REQUESTED":
		h.handlePasswordReset(ctx, ev, msg)
	default:
		log.Info().Str("event", ev.Event).Msg("unknown event, skipping")
		_ = msg.Ack(false)
	}
}

func (h *EventHandler) handleEmailVerification(ctx context.Context, ev model.NotificationEvent, msg amqp091.Delivery) {
	if !hasChannel(ev.Channels, "email") {
		_ = msg.Ack(false)
		return
	}

	email, _ := ev.Data["email"].(string)
	if email == "" {
		log.Warn().Msg("EMAIL_VERIFICATION_REQUESTED missing data.email")
		_ = msg.Ack(false)
		return
	}

	job := model.EmailJob{
		Email:    email,
		Template: "email_verification",
		Locale:   defaultLocale(ev.Locale),
		Subject:  subjectFor(ev.Event, defaultLocale(ev.Locale)),
		Data:     ev.Data,
		Meta: map[string]any{
			"event":   ev.Event,
			"user_id": ev.UserID.String(),
		},
	}

	if err := h.email.Publish(ctx, job); err != nil {
		log.Error().Err(err).Msg("publish email job failed")
		_ = msg.Nack(false, true)
		return
	}

	_ = msg.Ack(false)
}

func (h *EventHandler) handlePasswordReset(ctx context.Context, ev model.NotificationEvent, msg amqp091.Delivery) {
	if !hasChannel(ev.Channels, "email") {
		_ = msg.Ack(false)
		return
	}

	email, _ := ev.Data["email"].(string)
	if email == "" {
		log.Warn().Msg("PASSWORD_RESET_REQUESTED missing data.email")
		_ = msg.Ack(false)
		return
	}

	job := model.EmailJob{
		Email:    email,
		Template: "password_reset",
		Locale:   defaultLocale(ev.Locale),
		Subject:  subjectFor(ev.Event, defaultLocale(ev.Locale)),
		Data:     ev.Data,
		Meta: map[string]any{
			"event":   ev.Event,
			"user_id": ev.UserID.String(),
		},
	}

	if err := h.email.Publish(ctx, job); err != nil {
		log.Error().Err(err).Msg("publish email job failed")
		_ = msg.Nack(false, true)
		return
	}

	_ = msg.Ack(false)
}

func hasChannel(channels []string, want string) bool {
	if len(channels) == 0 {
		return true
	}
	for _, c := range channels {
		if c == want {
			return true
		}
	}
	return false
}

func defaultLocale(loc string) string {
	if loc == "" {
		return "ru"
	}
	return loc
}

func subjectFor(event, locale string) string {
	if locale == "en" {
		switch event {
		case "EMAIL_VERIFICATION_REQUESTED":
			return "Email verification"
		case "PASSWORD_RESET_REQUESTED":
			return "Password reset"
		default:
			return "Notification"
		}
	}

	switch event {
	case "EMAIL_VERIFICATION_REQUESTED":
		return "Подтверждение почты"
	case "PASSWORD_RESET_REQUESTED":
		return "Сброс пароля"
	default:
		return "Уведомление"
	}
}
