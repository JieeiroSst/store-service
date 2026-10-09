package domain

import (
	"fmt"
	"regexp"
	"strings"
)

const DefaultSensitivePattern = `(?i)(.*(password|passwd|secret|token|apikey|api_key|private_key|card_number|cvv).*|(.*_)?(otp|salt|pin|pwd)(_.*)?)`

type SensitivePolicy struct {
	pattern string
	re      *regexp.Regexp
}

func NewSensitivePolicy(pattern string) (SensitivePolicy, error) {
	if pattern == "" {
		pattern = DefaultSensitivePattern
	}
	re, err := regexp.Compile(`^(?:` + pattern + `)$`)
	if err != nil {
		return SensitivePolicy{}, fmt.Errorf("invalid sensitive field pattern: %w", err)
	}
	return SensitivePolicy{pattern: pattern, re: re}, nil
}

func (p SensitivePolicy) Pattern() string {
	return p.pattern
}

func (p SensitivePolicy) IsSensitive(field string) bool {
	for _, part := range strings.Split(field, ".") {
		if p.re.MatchString(part) {
			return true
		}
	}
	return false
}

func (p SensitivePolicy) CheckFields(fields []string) error {
	for _, f := range fields {
		if p.IsSensitive(f) {
			return Invalid("field %q is not allowed", f)
		}
	}
	return nil
}

func (p SensitivePolicy) Redact(doc Document) Document {
	if source, ok := p.redactValue(doc.Source).(map[string]any); ok {
		doc.Source = source
	}
	for field := range doc.Highlight {
		if p.IsSensitive(field) {
			delete(doc.Highlight, field)
		}
	}
	if len(doc.Highlight) == 0 {
		doc.Highlight = nil
	}
	return doc
}

func (p SensitivePolicy) RedactAll(docs []Document) []Document {
	for i := range docs {
		docs[i] = p.Redact(docs[i])
	}
	return docs
}

func (p SensitivePolicy) redactValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, inner := range val {
			if p.re.MatchString(k) {
				delete(val, k)
				continue
			}
			val[k] = p.redactValue(inner)
		}
		return val
	case []any:
		for i := range val {
			val[i] = p.redactValue(val[i])
		}
		return val
	}
	return v
}
