package template

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
)

type Engine struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
}

func NewEngine() *Engine {
	return &Engine{templates: make(map[string]*template.Template)}
}

func (e *Engine) Compile(tmpl *model.Template) error {
	t, err := template.New(tmpl.ID).Parse(tmpl.Body)
	if err != nil {
		return fmt.Errorf("compile template %s: %w", tmpl.Name, err)
	}
	e.mu.Lock()
	e.templates[tmpl.ID] = t
	e.mu.Unlock()
	return nil
}

func (e *Engine) Render(_ context.Context, templateID string, data map[string]interface{}) (string, error) {
	e.mu.RLock()
	t, ok := e.templates[templateID]
	e.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("template %s not compiled", templateID)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template %s: %w", templateID, err)
	}
	return buf.String(), nil
}

func (e *Engine) RenderInline(body string, data map[string]interface{}) (string, error) {
	t, err := template.New("inline").Parse(body)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *Engine) RenderSubject(subject string, data map[string]interface{}) (string, error) {
	t, err := template.New("subj").Parse(subject)
	if err != nil {
		return subject, nil // return as-is if unparseable
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return subject, nil
	}
	return strings.TrimSpace(buf.String()), nil
}

func (e *Engine) Evict(templateID string) {
	e.mu.Lock()
	delete(e.templates, templateID)
	e.mu.Unlock()
}
