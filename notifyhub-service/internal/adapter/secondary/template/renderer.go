package template

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"strings"
	"sync"
	texttemplate "text/template"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
)

type compiled struct {
	updatedAt time.Time
	subject   *texttemplate.Template
	body      func(buf *bytes.Buffer, data any) error
}

type Renderer struct {
	mu    sync.RWMutex
	cache map[string]*compiled
}

func New() *Renderer {
	return &Renderer{cache: make(map[string]*compiled)}
}

func NewRenderer() port.TemplateRenderer { return New() }

func (r *Renderer) Compile(t *model.Template) error {
	_, err := r.compile(t)
	return err
}

func (r *Renderer) compile(t *model.Template) (*compiled, error) {
	subj, err := texttemplate.New(t.ID + ":subject").Parse(t.Subject)
	if err != nil {
		return nil, fmt.Errorf("compile subject of template %s: %w", t.Name, err)
	}

	c := &compiled{updatedAt: t.UpdatedAt, subject: subj}
	if t.Channel == model.ChannelEmail {
		body, err := htmltemplate.New(t.ID).Parse(t.Body)
		if err != nil {
			return nil, fmt.Errorf("compile template %s: %w", t.Name, err)
		}
		c.body = func(buf *bytes.Buffer, data any) error { return body.Execute(buf, data) }
	} else {
		body, err := texttemplate.New(t.ID).Parse(t.Body)
		if err != nil {
			return nil, fmt.Errorf("compile template %s: %w", t.Name, err)
		}
		c.body = func(buf *bytes.Buffer, data any) error { return body.Execute(buf, data) }
	}

	r.mu.Lock()
	r.cache[t.ID] = c
	r.mu.Unlock()
	return c, nil
}

func (r *Renderer) Render(t *model.Template, data map[string]interface{}) (string, string, error) {
	r.mu.RLock()
	c, ok := r.cache[t.ID]
	r.mu.RUnlock()
	if !ok || !c.updatedAt.Equal(t.UpdatedAt) {
		var err error
		if c, err = r.compile(t); err != nil {
			return "", "", err
		}
	}

	var subj bytes.Buffer
	if err := c.subject.Execute(&subj, data); err != nil {
		return "", "", fmt.Errorf("render subject of template %s: %w", t.Name, err)
	}
	var body bytes.Buffer
	if err := c.body(&body, data); err != nil {
		return "", "", fmt.Errorf("render template %s: %w", t.Name, err)
	}
	return strings.TrimSpace(subj.String()), body.String(), nil
}

func (r *Renderer) Evict(templateID string) {
	r.mu.Lock()
	delete(r.cache, templateID)
	r.mu.Unlock()
}
