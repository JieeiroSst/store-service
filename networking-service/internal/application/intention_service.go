package application

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type IntentionService struct {
	store        port.IntentionStore
	blocker      *Blocker
	defaultAllow bool
}

func NewIntentionService(store port.IntentionStore, blocker *Blocker, cfg *config.Config) *IntentionService {
	return &IntentionService{store: store, blocker: blocker, defaultAllow: cfg.Intentions.DefaultAllow}
}

func (s *IntentionService) List(ctx context.Context, q QueryOptions) ([]domain.Intention, QueryMeta, error) {
	var out []domain.Intention
	m, err := s.blocker.Query(ctx, s.store, q, func() (uint64, error) {
		var (
			idx uint64
			err error
		)
		idx, out, err = s.store.IntentionList()
		return idx, err
	})
	return out, m, err
}

func (s *IntentionService) find(match func(domain.Intention) bool) (*domain.Intention, error) {
	_, all, err := s.store.IntentionList()
	if err != nil {
		return nil, err
	}
	for _, in := range all {
		if match(in) {
			return &in, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *IntentionService) Get(id string) (*domain.Intention, error) {
	return s.find(func(in domain.Intention) bool { return in.ID == id })
}

func (s *IntentionService) GetExact(src, dst string) (*domain.Intention, error) {
	return s.find(func(in domain.Intention) bool { return in.SourceName == src && in.DestinationName == dst })
}

func (s *IntentionService) Create(in domain.Intention) (string, error) {
	if err := in.Validate(); err != nil {
		return "", err
	}
	in.ID = newID()
	return in.ID, s.store.IntentionUpsert(in)
}

func (s *IntentionService) Update(id string, in domain.Intention) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	if err := in.Validate(); err != nil {
		return err
	}
	in.ID = id
	return s.store.IntentionUpsert(in)
}

func (s *IntentionService) UpsertExact(in domain.Intention) error {
	if err := in.Validate(); err != nil {
		return err
	}
	in.ID = newID()
	existing, err := s.GetExact(in.SourceName, in.DestinationName)
	switch {
	case err == nil:
		in.ID = existing.ID
	case !errors.Is(err, domain.ErrNotFound):
		return err
	}
	return s.store.IntentionUpsert(in)
}

func (s *IntentionService) Delete(id string) error {
	return s.store.IntentionDelete(id)
}

func (s *IntentionService) DeleteExact(src, dst string) error {
	in, err := s.GetExact(src, dst)
	if err != nil {
		return err
	}
	return s.store.IntentionDelete(in.ID)
}

func (s *IntentionService) Check(src, dst string) (bool, *domain.Intention, error) {
	if src == "" || dst == "" {
		return false, nil, domain.Invalid("source and destination are required")
	}
	in, err := s.find(func(in domain.Intention) bool { return in.Matches(src, dst) })
	switch {
	case err == nil:
		return in.Action == domain.IntentionAllow, in, nil
	case errors.Is(err, domain.ErrNotFound):
		return s.defaultAllow, nil, nil
	}
	return false, nil, err
}

func (s *IntentionService) Match(by, name string) ([]domain.Intention, error) {
	if by != "source" && by != "destination" {
		return nil, domain.Invalid("by must be 'source' or 'destination'")
	}
	_, all, err := s.store.IntentionList()
	if err != nil {
		return nil, err
	}
	out := []domain.Intention{}
	for _, in := range all {
		field := in.DestinationName
		if by == "source" {
			field = in.SourceName
		}
		if field == name || field == domain.Wildcard {
			out = append(out, in)
		}
	}
	return out, nil
}
