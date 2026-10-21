package application

import (
	"context"
	"sync"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
)

type fakeLLM struct {
	completion model.Completion
	deltas     []string
	err        error
	streamErr  error
	calls      []model.Question
}

func (f *fakeLLM) Complete(_ context.Context, q model.Question) (model.Completion, error) {
	f.calls = append(f.calls, q)
	return f.completion, f.err
}

func (f *fakeLLM) Stream(_ context.Context, q model.Question) (port.DeltaStream, error) {
	f.calls = append(f.calls, q)
	if f.err != nil {
		return nil, f.err
	}
	return &fakeStream{deltas: f.deltas, final: f.completion, err: f.streamErr}, nil
}

type fakeStream struct {
	deltas []string
	pos    int
	final  model.Completion
	err    error
	closed bool
}

func (s *fakeStream) Next() bool {
	if s.pos >= len(s.deltas) {
		return false
	}
	s.pos++
	return true
}

func (s *fakeStream) Current() string          { return s.deltas[s.pos-1] }
func (s *fakeStream) Result() model.Completion { return s.final }
func (s *fakeStream) Err() error               { return s.err }
func (s *fakeStream) Close() error             { s.closed = true; return nil }

type fakeHistory struct {
	mu       sync.Mutex
	turns    []model.Turn
	entries  []model.HistoryEntry
	saveErr  error
	listErr  error
	gotLimit int
	gotUser  string
}

func (f *fakeHistory) SaveTurn(_ context.Context, turn model.Turn) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.turns = append(f.turns, turn)
	return f.saveErr
}

func (f *fakeHistory) ListByUser(_ context.Context, userID string, limit int) ([]model.HistoryEntry, error) {
	f.gotUser, f.gotLimit = userID, limit
	return f.entries, f.listErr
}

func (f *fakeHistory) savedTurns() []model.Turn {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]model.Turn(nil), f.turns...)
}
