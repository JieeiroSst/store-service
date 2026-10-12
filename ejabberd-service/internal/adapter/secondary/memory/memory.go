package memory

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type us struct {
	user   string
	server string
}

type UserStore struct {
	mu    sync.RWMutex
	users map[us]domain.Password
}

func NewUserStore() *UserStore {
	return &UserStore{users: map[us]domain.Password{}}
}

var _ port.UserRepository = (*UserStore)(nil)

func (s *UserStore) GetPassword(_ context.Context, luser, lserver string) (domain.Password, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.users[us{luser, lserver}]
	if !ok {
		return domain.Password{}, domain.ErrUserNotFound
	}
	return p, nil
}

func (s *UserStore) SetPassword(_ context.Context, luser, lserver string, pass domain.Password) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[us{luser, lserver}] = pass
	return nil
}

func (s *UserStore) TryRegister(_ context.Context, luser, lserver string, pass domain.Password) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := us{luser, lserver}
	if _, ok := s.users[k]; ok {
		return domain.ErrUserExists
	}
	s.users[k] = pass
	return nil
}

func (s *UserStore) UserExists(_ context.Context, luser, lserver string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.users[us{luser, lserver}]
	return ok, nil
}

func (s *UserStore) RemoveUser(_ context.Context, luser, lserver string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, us{luser, lserver})
	return nil
}

func (s *UserStore) ListUsers(_ context.Context, lserver string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []string
	for k := range s.users {
		if k.server == lserver {
			out = append(out, k.user)
		}
	}
	slices.Sort(out)
	return out, nil
}

func (s *UserStore) CountUsers(ctx context.Context, lserver string) (int, error) {
	users, err := s.ListUsers(ctx, lserver)
	return len(users), err
}

type RosterStore struct {
	mu    sync.Mutex
	items map[us]map[domain.LJID]domain.RosterItem
}

func NewRosterStore() *RosterStore {
	return &RosterStore{items: map[us]map[domain.LJID]domain.RosterItem{}}
}

var _ port.RosterRepository = (*RosterStore)(nil)

func copyItem(it domain.RosterItem) domain.RosterItem {
	it.Groups = slices.Clone(it.Groups)
	xs := make([]*domain.Element, len(it.Xs))
	for i, x := range it.Xs {
		xs[i] = x.Copy()
	}
	it.Xs = xs
	return it
}

func (s *RosterStore) GetRoster(_ context.Context, luser, lserver string) ([]domain.RosterItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.list(luser, lserver), nil
}

func (s *RosterStore) list(luser, lserver string) []domain.RosterItem {
	var out []domain.RosterItem
	for _, it := range s.items[us{luser, lserver}] {
		out = append(out, copyItem(it))
	}
	slices.SortFunc(out, func(a, b domain.RosterItem) int {
		switch {
		case a.JID.String() < b.JID.String():
			return -1
		case a.JID.String() > b.JID.String():
			return 1
		}
		return 0
	})
	return out
}

func (s *RosterStore) GetRosterItem(_ context.Context, luser, lserver string, jid domain.LJID) (domain.RosterItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.get(luser, lserver, jid)
}

func (s *RosterStore) get(luser, lserver string, jid domain.LJID) (domain.RosterItem, bool, error) {
	it, ok := s.items[us{luser, lserver}][jid]
	if !ok {
		return domain.RosterItem{}, false, nil
	}
	return copyItem(it), true, nil
}

func (s *RosterStore) UpdateRoster(_ context.Context, item domain.RosterItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.put(item)
	return nil
}

func (s *RosterStore) put(item domain.RosterItem) {
	k := us{item.LUser, item.LServer}
	if s.items[k] == nil {
		s.items[k] = map[domain.LJID]domain.RosterItem{}
	}
	s.items[k][item.JID] = copyItem(item)
}

func (s *RosterStore) DeleteRoster(_ context.Context, luser, lserver string, jid domain.LJID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items[us{luser, lserver}], jid)
	return nil
}

func (s *RosterStore) RemoveUser(_ context.Context, luser, lserver string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, us{luser, lserver})
	return nil
}

func (s *RosterStore) Transaction(_ context.Context, _, _ string, fn func(tx port.RosterRepository) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn(&rosterTx{s: s})
}

type rosterTx struct {
	s *RosterStore
}

func (t *rosterTx) GetRoster(_ context.Context, luser, lserver string) ([]domain.RosterItem, error) {
	return t.s.list(luser, lserver), nil
}

func (t *rosterTx) GetRosterItem(_ context.Context, luser, lserver string, jid domain.LJID) (domain.RosterItem, bool, error) {
	return t.s.get(luser, lserver, jid)
}

func (t *rosterTx) UpdateRoster(_ context.Context, item domain.RosterItem) error {
	t.s.put(item)
	return nil
}

func (t *rosterTx) DeleteRoster(_ context.Context, luser, lserver string, jid domain.LJID) error {
	delete(t.s.items[us{luser, lserver}], jid)
	return nil
}

func (t *rosterTx) RemoveUser(_ context.Context, luser, lserver string) error {
	delete(t.s.items, us{luser, lserver})
	return nil
}

func (t *rosterTx) Transaction(ctx context.Context, _, _ string, fn func(tx port.RosterRepository) error) error {
	return fn(t)
}

type OfflineStore struct {
	mu   sync.Mutex
	msgs map[us][]domain.OfflineMessage
}

func NewOfflineStore() *OfflineStore {
	return &OfflineStore{msgs: map[us][]domain.OfflineMessage{}}
}

var _ port.OfflineRepository = (*OfflineStore)(nil)

func (s *OfflineStore) Store(_ context.Context, msg domain.OfflineMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := us{msg.LUser, msg.LServer}
	msg.Packet = msg.Packet.Copy()
	s.msgs[k] = append(s.msgs[k], msg)
	return nil
}

func (s *OfflineStore) Pop(_ context.Context, luser, lserver string) ([]domain.OfflineMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := us{luser, lserver}
	out := s.msgs[k]
	delete(s.msgs, k)
	return out, nil
}

func (s *OfflineStore) Count(_ context.Context, luser, lserver string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.msgs[us{luser, lserver}]), nil
}

func (s *OfflineStore) RemoveUser(_ context.Context, luser, lserver string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.msgs, us{luser, lserver})
	return nil
}

func (s *OfflineStore) RemoveExpired(_ context.Context, lserver string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	n := 0
	for k, list := range s.msgs {
		if k.server != lserver {
			continue
		}
		kept := list[:0]
		for _, m := range list {
			if m.Expire != nil && !m.Expire.After(now) {
				n++
				continue
			}
			kept = append(kept, m)
		}
		s.msgs[k] = kept
	}
	return n, nil
}
