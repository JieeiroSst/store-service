package faq

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type faqFile struct {
	FAQs []struct {
		Question string `yaml:"question"`
		Answer   string `yaml:"answer"`
	} `yaml:"faqs"`
}

func LoadFromFile(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("faq: failed to read %s: %w", path, err)
	}

	var parsed faqFile
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("faq: failed to parse %s: %w", path, err)
	}

	entries := make([]Entry, 0, len(parsed.FAQs))
	for _, f := range parsed.FAQs {
		if f.Question == "" || f.Answer == "" {
			continue
		}
		entries = append(entries, Entry{Question: f.Question, Answer: f.Answer})
	}
	return entries, nil
}
