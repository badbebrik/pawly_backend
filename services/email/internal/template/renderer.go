package template

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

type Renderer struct {
	baseDir string
}

func NewRenderer(baseDir string) *Renderer {
	return &Renderer{baseDir: baseDir}
}

func (r *Renderer) Render(locale, tplName string, data map[string]any) (string, error) {
	if locale == "" {
		locale = "ru"
	}

	path := filepath.Join(r.baseDir, locale, tplName+".html")

	tpl, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", path, err)
	}

	return buf.String(), nil
}
