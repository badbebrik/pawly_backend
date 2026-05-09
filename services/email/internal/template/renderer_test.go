package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRendererRenderUsesLocaleTemplate(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "ru")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "welcome_email.html"), []byte("Hello {{.name}}"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	html, err := NewRenderer(dir).Render("ru", "welcome_email", map[string]any{"name": "Ivan"})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if html != "Hello Ivan" {
		t.Fatalf("unexpected html: %q", html)
	}
}

func TestRendererRenderReturnsParseError(t *testing.T) {
	_, err := NewRenderer(t.TempDir()).Render("ru", "missing", nil)
	if err == nil || !strings.Contains(err.Error(), "parse template") {
		t.Fatalf("expected parse template error, got %v", err)
	}
}
