package faq

import (
	"context"
	"fmt"
	"sync"
)

type VoyageEmbedder interface {
	Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error)
}

type Store struct {
	mu        sync.RWMutex
	entries   []Entry
	threshold float64
}

func NewStore(threshold float64) *Store {
	return &Store{threshold: threshold}
}

func (s *Store) Load(ctx context.Context, embedder VoyageEmbedder, entries []Entry) error {
	if len(entries) == 0 {
		s.mu.Lock()
		s.entries = nil
		s.mu.Unlock()
		return nil
	}

	questions := make([]string, len(entries))
	for i, e := range entries {
		questions[i] = e.Question
	}

	embeddings, err := embedder.Embed(ctx, questions, "document")
	if err != nil {
		return fmt.Errorf("faq: failed to embed FAQ entries: %w", err)
	}
	if len(embeddings) != len(entries) {
		return fmt.Errorf("faq: embedder returned %d embeddings for %d entries", len(embeddings), len(entries))
	}

	loaded := make([]Entry, len(entries))
	for i, e := range entries {
		e.Embedding = embeddings[i]
		loaded[i] = e
	}

	s.mu.Lock()
	s.entries = loaded
	s.mu.Unlock()
	return nil
}

func (s *Store) Match(ctx context.Context, embedder VoyageEmbedder, question string) (entry Entry, score float64, ok bool, err error) {
	s.mu.RLock()
	entries := s.entries
	s.mu.RUnlock()

	if len(entries) == 0 {
		return Entry{}, 0, false, nil
	}

	embeddings, err := embedder.Embed(ctx, []string{question}, "query")
	if err != nil {
		return Entry{}, 0, false, err
	}
	if len(embeddings) == 0 {
		return Entry{}, 0, false, fmt.Errorf("faq: embedder returned no embedding for question")
	}
	queryEmbedding := embeddings[0]

	var best Entry
	var bestScore float64
	for _, e := range entries {
		score := CosineSimilarity(queryEmbedding, e.Embedding)
		if score > bestScore {
			bestScore = score
			best = e
		}
	}

	if bestScore >= s.threshold {
		return best, bestScore, true, nil
	}
	return Entry{}, bestScore, false, nil
}
