package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
)

type Challenges struct {
	mu   sync.RWMutex
	byID map[string]*domain.Challenge
}

func NewChallenges() *Challenges {
	return &Challenges{byID: map[string]*domain.Challenge{}}
}

func (r *Challenges) Create(_ context.Context, c *domain.Challenge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[c.ID]; ok {
		return domain.Conflict("challenge %s already exists", c.ID)
	}
	r.byID[c.ID] = c.Clone()
	return nil
}

func (r *Challenges) Update(_ context.Context, c *domain.Challenge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[c.ID]; !ok {
		return domain.ErrNotFound
	}
	r.byID[c.ID] = c.Clone()
	return nil
}

func (r *Challenges) Get(_ context.Context, id string) (*domain.Challenge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c.Clone(), nil
}

func (r *Challenges) ListPendingByUser(_ context.Context, userID int64, now time.Time) ([]*domain.Challenge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []*domain.Challenge{}
	for _, c := range r.byID {
		if c.UserID == userID && c.Status == domain.StatusPending && now.Before(c.ExpiresAt) {
			out = append(out, c.Clone())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

type Locker struct {
	mu    sync.Mutex
	locks map[string]*keyLock
}

type keyLock struct {
	mu   sync.Mutex
	refs int
}

func NewLocker() *Locker {
	return &Locker{locks: map[string]*keyLock{}}
}

func (l *Locker) WithLock(ctx context.Context, key string, fn func(ctx context.Context) error) error {
	l.mu.Lock()
	k, ok := l.locks[key]
	if !ok {
		k = &keyLock{}
		l.locks[key] = k
	}
	k.refs++
	l.mu.Unlock()

	k.mu.Lock()
	defer func() {
		k.mu.Unlock()
		l.mu.Lock()
		if k.refs--; k.refs == 0 {
			delete(l.locks, key)
		}
		l.mu.Unlock()
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(ctx)
}
