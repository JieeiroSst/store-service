package memory

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

type Customers struct {
	mu     sync.RWMutex
	byID   map[string]*domain.Customer
	byUser map[int64]string
}

func NewCustomers() *Customers {
	return &Customers{byID: map[string]*domain.Customer{}, byUser: map[int64]string{}}
}

func (r *Customers) Create(_ context.Context, c *domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byUser[c.UserID]; ok {
		return domain.Conflict("user %d is already a customer", c.UserID)
	}
	r.byID[c.ID] = c.Clone()
	r.byUser[c.UserID] = c.ID
	return nil
}

func (r *Customers) Update(_ context.Context, c *domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[c.ID]; !ok {
		return domain.ErrNotFound
	}
	if c.KYC.Status == domain.KYCVerified && c.KYC.DocumentHash != "" {
		for id, other := range r.byID {
			if id != c.ID && other.KYC.Status == domain.KYCVerified && other.KYC.DocumentHash == c.KYC.DocumentHash {
				return domain.Conflict("this identity document is already verified for another customer")
			}
		}
	}
	r.byID[c.ID] = c.Clone()
	return nil
}

func (r *Customers) Get(_ context.Context, id string) (*domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c.Clone(), nil
}

func (r *Customers) GetByUserID(ctx context.Context, userID int64) (*domain.Customer, error) {
	r.mu.RLock()
	id, ok := r.byUser[userID]
	r.mu.RUnlock()
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r.Get(ctx, id)
}

func (r *Customers) GetByDocumentHash(_ context.Context, hash string) (*domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.byID {
		if c.KYC.Status == domain.KYCVerified && c.KYC.DocumentHash == hash {
			return c.Clone(), nil
		}
	}
	return nil, domain.ErrNotFound
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
