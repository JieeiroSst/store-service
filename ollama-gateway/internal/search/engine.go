package search

import (
	"fmt"
	"strings"

	"ollama-gateway/internal/preset"
)

type Result struct {
	QA    preset.QA
	Score float64
}

type Engine struct {
	index []indexedDoc
}

type indexedDoc struct {
	qa   preset.QA
	text string // normalized searchable text
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Build(items []preset.QA) {
	e.index = make([]indexedDoc, len(items))
	for i, qa := range items {
		parts := []string{
			normalize(qa.Question),
			normalize(qa.Answer[:min(len(qa.Answer), 200)]),
		}
		for _, kw := range qa.Keywords {
			parts = append(parts, normalize(kw))
		}
		e.index[i] = indexedDoc{
			qa:   qa,
			text: strings.Join(parts, " "),
		}
	}
}

func (e *Engine) Search(query string, threshold float64) *Result {
	if len(e.index) == 0 {
		return nil
	}

	queryTerms := tokenize(normalize(query))
	if len(queryTerms) == 0 {
		return nil
	}

	var best *Result

	for _, doc := range e.index {
		docTerms := tokenize(doc.text)
		score := tfScore(queryTerms, docTerms)

		if score > threshold {
			if best == nil || score > best.Score {
				best = &Result{QA: doc.qa, Score: score}
			}
		}
	}

	return best
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func tokenize(s string) []string {
	// Split on whitespace and common punctuation
	replacer := strings.NewReplacer(
		"?", " ", "!", " ", ".", " ", ",", " ",
		":", " ", ";", " ", "(", " ", ")", " ",
	)
	s = replacer.Replace(s)
	raw := strings.Fields(s)

	terms := make([]string, 0, len(raw))
	for _, t := range raw {
		if len(t) > 1 { // skip single chars
			terms = append(terms, t)
		}
	}
	return terms
}

func tfScore(queryTerms, docTerms []string) float64 {
	if len(queryTerms) == 0 || len(docTerms) == 0 {
		return 0
	}

	docFreq := make(map[string]int, len(docTerms))
	for _, t := range docTerms {
		docFreq[t]++
	}

	matches := 0
	for _, qt := range queryTerms {
		if docFreq[qt] > 0 {
			matches++
		}
		for dt := range docFreq {
			if len(qt) >= 3 && strings.HasPrefix(dt, qt) {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(queryTerms))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (e *Engine) DebugIndex() string {
	return fmt.Sprintf("Search index: %d documents", len(e.index))
}
