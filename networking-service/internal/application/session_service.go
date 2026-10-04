package application

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type SessionService struct {
	store    port.SessionStore
	blocker  *Blocker
	metrics  port.Metrics
	selfNode string
	now      func() time.Time
}

func NewSessionService(store port.SessionStore, blocker *Blocker, metrics port.Metrics, catalog *CatalogService) *SessionService {
	return &SessionService{
		store:    store,
		blocker:  blocker,
		metrics:  metrics,
		selfNode: catalog.Self().NodeName,
		now:      time.Now,
	}
}

func (s *SessionService) Create(sess domain.Session) (string, error) {
	if sess.Node == "" {
		sess.Node = s.selfNode
	}
	if err := sess.Validate(); err != nil {
		return "", err
	}
	sess.ID = newID()
	if sess.TTL > 0 {
		sess.Expires = s.now().Add(sess.TTL)
	}
	if err := s.store.SessionCreate(sess); err != nil {
		return "", err
	}
	return sess.ID, nil
}

func (s *SessionService) Renew(id string) (*domain.Session, error) {
	_, sess, err := s.store.SessionGet(id)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, domain.ErrNotFound
	}
	var expires time.Time
	if sess.TTL > 0 {
		expires = s.now().Add(sess.TTL)
	}
	return s.store.SessionRenew(id, expires)
}

func (s *SessionService) Destroy(id string) error {
	err := s.store.SessionDestroy(id, s.now())
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err == nil {
		s.metrics.SessionInvalidated("destroyed")
	}
	return err
}

func (s *SessionService) Info(ctx context.Context, q QueryOptions, id string) (*domain.Session, QueryMeta, error) {
	var out *domain.Session
	m, err := s.blocker.Query(ctx, s.store, q, func() (uint64, error) {
		var (
			idx uint64
			err error
		)
		idx, out, err = s.store.SessionGet(id)
		return idx, err
	})
	return out, m, err
}

func (s *SessionService) List(ctx context.Context, q QueryOptions, node string) ([]domain.Session, QueryMeta, error) {
	var out []domain.Session
	m, err := s.blocker.Query(ctx, s.store, q, func() (uint64, error) {
		idx, all, err := s.store.SessionList()
		out = all[:0]
		for _, sess := range all {
			if node == "" || sess.Node == node {
				out = append(out, sess)
			}
		}
		return idx, err
	})
	return out, m, err
}

func (s *SessionService) ReapExpired() error {
	now := s.now()
	_, all, err := s.store.SessionList()
	if err != nil {
		return err
	}
	for _, sess := range all {
		if sess.TTL == 0 || sess.Expires.IsZero() || !now.After(sess.Expires.Add(sess.TTL)) {
			continue
		}
		if err := s.store.SessionDestroy(sess.ID, now); err == nil {
			s.metrics.SessionInvalidated("ttl")
		}
	}
	return nil
}
