package handler

import (
	"email/internal/model"
	"testing"
)

func TestMapEventToEmailJobCreatesVerificationJob(t *testing.T) {
	job, ok := mapEventToEmailJob(model.NotificationEvent{
		Event:    "EMAIL_VERIFICATION_REQUESTED",
		UserID:   "user-1",
		Locale:   "en",
		Channels: []string{"email"},
		Data: map[string]any{
			"email": "ivan@example.com",
			"code":  "123456",
		},
	})
	if !ok {
		t.Fatalf("expected event to map to job")
	}
	if job.Email != "ivan@example.com" || job.Template != "email_verification" || job.Locale != "en" {
		t.Fatalf("unexpected job: %+v", job)
	}
	if job.Subject != "Your Pawly verification code" {
		t.Fatalf("unexpected subject: %q", job.Subject)
	}
}

func TestMapEventToEmailJobSkipsUnsupportedEvents(t *testing.T) {
	if _, ok := mapEventToEmailJob(model.NotificationEvent{
		Event:    "UNKNOWN",
		Channels: []string{"email"},
		Data:     map[string]any{"email": "ivan@example.com"},
	}); ok {
		t.Fatalf("expected unknown event to be skipped")
	}

	if _, ok := mapEventToEmailJob(model.NotificationEvent{
		Event:    "WELCOME_EMAIL",
		Channels: []string{"push"},
		Data:     map[string]any{"email": "ivan@example.com"},
	}); ok {
		t.Fatalf("expected non-email channel to be skipped")
	}

	if _, ok := mapEventToEmailJob(model.NotificationEvent{
		Event: "WELCOME_EMAIL",
		Data:  map[string]any{"email": " "},
	}); ok {
		t.Fatalf("expected missing email to be skipped")
	}
}
