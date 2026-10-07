package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
)

type Store struct {
	mu       sync.RWMutex
	accounts map[string]*domain.Account
	cards    map[string]*domain.Card
	panIndex map[string]string
	auths    map[string]*domain.Authorization
	byCard   map[string][]string
	txns     map[string][]*domain.Transaction
}

func NewStore() *Store {
	return &Store{
		accounts: map[string]*domain.Account{},
		cards:    map[string]*domain.Card{},
		panIndex: map[string]string{},
		auths:    map[string]*domain.Authorization{},
		byCard:   map[string][]string{},
		txns:     map[string][]*domain.Transaction{},
	}
}

type Accounts struct{ s *Store }

func NewAccounts(s *Store) Accounts { return Accounts{s: s} }

func (r Accounts) Create(_ context.Context, a *domain.Account) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *a
	r.s.accounts[a.ID] = &cp
	return nil
}

func (r Accounts) Update(_ context.Context, a *domain.Account) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.accounts[a.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *a
	r.s.accounts[a.ID] = &cp
	return nil
}

func (r Accounts) Get(_ context.Context, id string) (*domain.Account, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	a, ok := r.s.accounts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r Accounts) ListByCustomer(_ context.Context, customerID string) ([]*domain.Account, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []*domain.Account{}
	for _, a := range r.s.accounts {
		if a.CustomerID == customerID {
			cp := *a
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

type Cards struct{ s *Store }

func NewCards(s *Store) Cards { return Cards{s: s} }

func (r Cards) Create(_ context.Context, c *domain.Card) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.panIndex[c.PANHash]; ok {
		return domain.ErrDuplicatePAN
	}
	r.s.cards[c.ID] = c.Clone()
	r.s.panIndex[c.PANHash] = c.ID
	return nil
}

func (r Cards) Update(_ context.Context, c *domain.Card) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.cards[c.ID]; !ok {
		return domain.ErrNotFound
	}
	r.s.cards[c.ID] = c.Clone()
	return nil
}

func (r Cards) Get(_ context.Context, id string) (*domain.Card, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.cards[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c.Clone(), nil
}

func (r Cards) GetByPANHash(ctx context.Context, hash string) (*domain.Card, error) {
	r.s.mu.RLock()
	id, ok := r.s.panIndex[hash]
	r.s.mu.RUnlock()
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r.Get(ctx, id)
}

func (r Cards) ListByAccount(_ context.Context, accountID string) ([]*domain.Card, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []*domain.Card{}
	for _, c := range r.s.cards {
		if c.AccountID == accountID {
			out = append(out, c.Clone())
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

type Authorizations struct{ s *Store }

func NewAuthorizations(s *Store) Authorizations { return Authorizations{s: s} }

func (r Authorizations) Create(_ context.Context, a *domain.Authorization) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *a
	r.s.auths[a.ID] = &cp
	r.s.byCard[a.CardID] = append(r.s.byCard[a.CardID], a.ID)
	return nil
}

func (r Authorizations) Update(_ context.Context, a *domain.Authorization) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.auths[a.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *a
	r.s.auths[a.ID] = &cp
	return nil
}

func (r Authorizations) Get(_ context.Context, id string) (*domain.Authorization, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	a, ok := r.s.auths[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (r Authorizations) ListByCard(_ context.Context, cardID string, limit int) ([]*domain.Authorization, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	ids := r.s.byCard[cardID]
	out := []*domain.Authorization{}
	for i := len(ids) - 1; i >= 0 && len(out) < limit; i-- {
		cp := *r.s.auths[ids[i]]
		out = append(out, &cp)
	}
	return out, nil
}

func (r Authorizations) SumDebits(_ context.Context, cardID string, since time.Time) (int64, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var sum int64
	for _, id := range r.s.byCard[cardID] {
		a := r.s.auths[id]
		if !a.ProcessingCode.Debit() || a.CreatedAt.Before(since) {
			continue
		}
		switch a.Status {
		case domain.AuthAuthorized:
			sum += a.Amount
		case domain.AuthConfirmed:
			sum += a.ConfirmedAmount
		}
	}
	return sum, nil
}

type Transactions struct{ s *Store }

func NewTransactions(s *Store) Transactions { return Transactions{s: s} }

func (r Transactions) Create(_ context.Context, t *domain.Transaction) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *t
	r.s.txns[t.AccountID] = append(r.s.txns[t.AccountID], &cp)
	return nil
}

func (r Transactions) ListByAccount(_ context.Context, accountID string, limit int) ([]*domain.Transaction, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	list := r.s.txns[accountID]
	out := []*domain.Transaction{}
	for i := len(list) - 1; i >= 0 && len(out) < limit; i-- {
		cp := *list[i]
		out = append(out, &cp)
	}
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
