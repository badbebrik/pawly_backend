package service

import (
	"context"
	"email/internal/model"
	"email/internal/smtp"
	emailtemplate "email/internal/template"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDispatcherProcessSendsRenderedEmail(t *testing.T) {
	restoreBackoff := setEmailTestBackoff()
	defer restoreBackoff()

	renderer := newEmailTestRenderer(t)
	primary := &stubEmailProvider{name: "primary"}
	dispatcher := NewDispatcher(renderer, primary, nil, false)

	err := dispatcher.Process(context.Background(), model.EmailJob{
		Email:    "ivan@example.com",
		Template: "welcome_email",
		Locale:   "ru",
		Subject:  "Welcome",
		Data:     map[string]any{"name": "Ivan"},
	})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}

	if len(primary.messages) != 1 {
		t.Fatalf("expected one sent message, got %d", len(primary.messages))
	}
	if primary.messages[0].To != "ivan@example.com" || primary.messages[0].Subject != "Welcome" {
		t.Fatalf("unexpected message: %+v", primary.messages[0])
	}
	if primary.messages[0].HTML != "Hello Ivan" {
		t.Fatalf("unexpected html: %q", primary.messages[0].HTML)
	}
}

func TestDispatcherProcessSkipsInvalidJob(t *testing.T) {
	restoreBackoff := setEmailTestBackoff()
	defer restoreBackoff()

	renderer := newEmailTestRenderer(t)
	primary := &stubEmailProvider{name: "primary"}
	dispatcher := NewDispatcher(renderer, primary, nil, false)

	if err := dispatcher.Process(context.Background(), model.EmailJob{Template: "welcome_email"}); err != nil {
		t.Fatalf("expected invalid job to be skipped, got %v", err)
	}
	if len(primary.messages) != 0 {
		t.Fatalf("did not expect provider calls, got %d", len(primary.messages))
	}
}

func TestDispatcherProcessUsesFallbackProvider(t *testing.T) {
	restoreBackoff := setEmailTestBackoff()
	defer restoreBackoff()

	renderer := newEmailTestRenderer(t)
	primary := &stubEmailProvider{name: "primary", err: errors.New("primary failed")}
	fallback := &stubEmailProvider{name: "fallback"}
	dispatcher := NewDispatcher(renderer, primary, fallback, true)

	err := dispatcher.Process(context.Background(), model.EmailJob{
		Email:    "ivan@example.com",
		Template: "welcome_email",
		Locale:   "ru",
		Subject:  "Welcome",
		Data:     map[string]any{"name": "Ivan"},
	})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}

	if len(primary.messages) != 1 || len(fallback.messages) != 1 {
		t.Fatalf("expected primary and fallback calls, primary=%d fallback=%d", len(primary.messages), len(fallback.messages))
	}
	if !dispatcher.RequeueOnFail() {
		t.Fatalf("expected requeue flag")
	}
}

func TestDispatcherProcessReturnsRenderError(t *testing.T) {
	restoreBackoff := setEmailTestBackoff()
	defer restoreBackoff()

	renderer := newEmailTestRenderer(t)
	primary := &stubEmailProvider{name: "primary"}
	dispatcher := NewDispatcher(renderer, primary, nil, false)

	err := dispatcher.Process(context.Background(), model.EmailJob{
		Email:    "ivan@example.com",
		Template: "missing_template",
		Locale:   "ru",
		Subject:  "Welcome",
	})
	if err == nil {
		t.Fatalf("expected render error")
	}
	if len(primary.messages) != 0 {
		t.Fatalf("did not expect provider calls, got %d", len(primary.messages))
	}
}

func newEmailTestRenderer(t *testing.T) *emailtemplate.Renderer {
	t.Helper()
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "ru")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "welcome_email.html"), []byte("Hello {{.name}}"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	return emailtemplate.NewRenderer(dir)
}

func setEmailTestBackoff() func() {
	old := BackoffSchedule
	BackoffSchedule = []time.Duration{0}
	return func() {
		BackoffSchedule = old
	}
}

type stubEmailProvider struct {
	name     string
	err      error
	messages []smtp.Message
}

func (p *stubEmailProvider) Send(_ context.Context, msg smtp.Message) error {
	p.messages = append(p.messages, msg)
	return p.err
}

func (p *stubEmailProvider) Name() string {
	return p.name
}

var _ smtp.Provider = (*stubEmailProvider)(nil)
