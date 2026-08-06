package template

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

//go:embed templates/*.html
var templatesFS embed.FS

// Each template file defines two named blocks, "subject" and "html", both
// executed against the same map[string]string data:
//
//	{{define "subject"}}...{{end}}
//	{{define "html"}}...{{end}}
type renderer struct {
	templates map[string]*template.Template
}

func NewRenderer() (port.TemplateRenderer, error) {
	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("read email templates dir: %w", err)
	}

	templates := make(map[string]*template.Template, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".html")
		tmpl, err := template.ParseFS(templatesFS, "templates/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("parse email template %q: %w", entry.Name(), err)
		}
		templates[name] = tmpl
	}

	return &renderer{templates: templates}, nil
}

func (r *renderer) Render(templateType string, data map[string]string) (string, string, error) {
	tmpl, ok := r.templates[templateType]
	if !ok {
		return "", "", fmt.Errorf("unknown email template %q", templateType)
	}

	var subjectBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&subjectBuf, "subject", data); err != nil {
		return "", "", fmt.Errorf("render subject for template %q: %w", templateType, err)
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&htmlBuf, "html", data); err != nil {
		return "", "", fmt.Errorf("render html for template %q: %w", templateType, err)
	}

	return strings.TrimSpace(subjectBuf.String()), htmlBuf.String(), nil
}
