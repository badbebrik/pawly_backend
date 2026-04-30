package handler

import (
	"context"
	"encoding/json"
	"strings"

	"email/internal/model"
	"email/internal/service"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type EventHandler struct {
	dispatcher *service.Dispatcher
}

func NewEventHandler(dispatcher *service.Dispatcher) *EventHandler {
	return &EventHandler{dispatcher: dispatcher}
}

func (h *EventHandler) Handle(ctx context.Context, msg amqp091.Delivery) {
	var event model.NotificationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Warn().Err(err).Msg("invalid notification event json")
		_ = msg.Ack(false)
		return
	}

	job, ok := mapEventToEmailJob(event)
	if !ok {
		_ = msg.Ack(false)
		return
	}

	if err := h.dispatcher.Process(ctx, job); err != nil {
		log.Error().Err(err).Str("event", event.Event).Msg("process notification email failed")
		if h.dispatcher.RequeueOnFail() {
			_ = msg.Nack(false, true)
			return
		}
		_ = msg.Nack(false, false)
		return
	}

	_ = msg.Ack(false)
}

func mapEventToEmailJob(event model.NotificationEvent) (model.EmailJob, bool) {
	if !hasChannel(event.Channels, "email") {
		return model.EmailJob{}, false
	}

	email, _ := event.Data["email"].(string)
	if strings.TrimSpace(email) == "" {
		log.Warn().Str("event", event.Event).Msg("notification event missing data.email")
		return model.EmailJob{}, false
	}

	switch event.Event {
	case "EMAIL_VERIFICATION_REQUESTED":
		return newEmailJob(event, email, "email_verification"), true
	case "PASSWORD_RESET_REQUESTED":
		return newEmailJob(event, email, "password_reset"), true
	case "WELCOME_EMAIL":
		return newEmailJob(event, email, "welcome_email"), true
	default:
		log.Info().Str("event", event.Event).Msg("unknown notification event, skipping")
		return model.EmailJob{}, false
	}
}

func newEmailJob(event model.NotificationEvent, email, template string) model.EmailJob {
	return model.EmailJob{
		Email:    email,
		Template: template,
		Locale:   defaultLocale(event.Locale),
		Subject:  subjectFor(event.Event, defaultLocale(event.Locale)),
		Data:     event.Data,
		Meta: map[string]any{
			"event":   event.Event,
			"user_id": event.UserID,
		},
	}
}

func hasChannel(channels []string, want string) bool {
	if len(channels) == 0 {
		return true
	}
	for i := range channels {
		if channels[i] == want {
			return true
		}
	}
	return false
}

func defaultLocale(locale string) string {
	if strings.TrimSpace(locale) == "" {
		return "ru"
	}
	return locale
}

func subjectFor(event, locale string) string {
	if locale == "en" {
		switch event {
		case "EMAIL_VERIFICATION_REQUESTED":
			return "Your Pawly verification code"
		case "PASSWORD_RESET_REQUESTED":
			return "Your Pawly password reset code"
		case "WELCOME_EMAIL":
			return "Добро пожаловать в Pawly"
		default:
			return "Notification"
		}
	}

	switch event {
	case "EMAIL_VERIFICATION_REQUESTED":
		return "Код подтверждения Pawly"
	case "PASSWORD_RESET_REQUESTED":
		return "Код сброса пароля Pawly"
	case "WELCOME_EMAIL":
		return "Добро пожаловать в Pawly"
	default:
		return "Уведомление"
	}
}
