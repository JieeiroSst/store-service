package slacktemplate

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

//go:embed templates/*.txt
var templatesFS embed.FS

// Each template file defines two named blocks, "title" and "text", both
// executed against the same map[string]string data. Plain text/template
// (not html/template) is used since Slack mrkdwn is not HTML and shouldn't
// be HTML-escaped:
//
//	{{define "title"}}...{{end}}
//	{{define "text"}}...{{end}}
type renderer struct {
	templates map[string]*template.Template
}

func NewRenderer() (port.SlackTemplateRenderer, error) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("read slack templates dir: %w", err)
	}

	templates := make(map[string]*template.Template, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".txt")
		tmpl, err := template.ParseFS(templatesFS, "templates/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("parse slack template %q: %w", entry.Name(), err)
		}
		templates[name] = tmpl
	}

	return &renderer{templates: templates}, nil
}

func (r *renderer) Render(templateType string, data map[string]string) (string, string, error) {
	tmpl, ok := r.templates[templateType]
	if !ok {
		return "", "", fmt.Errorf("unknown slack template %q", templateType)
	}

	var titleBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&titleBuf, "title", data); err != nil {
		return "", "", fmt.Errorf("render title for template %q: %w", templateType, err)
	}

	var textBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&textBuf, "text", data); err != nil {
		return "", "", fmt.Errorf("render text for template %q: %w", templateType, err)
	}

	return strings.TrimSpace(titleBuf.String()), strings.TrimSpace(textBuf.String()), nil
}
