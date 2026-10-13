package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
)

type Repository struct {
	mu        sync.Mutex
	nextID    int64
	wishlists map[int64]*model.Wishlist
	Conflicts int
}

func New() *Repository {
	return &Repository{wishlists: map[int64]*model.Wishlist{}}
}

func clone(w *model.Wishlist) *model.Wishlist {
	c := *w
	c.Lines = append([]model.Line{}, w.Lines...)
	return &c
}

func (m *Repository) Name() string { return "memory" }

func (m *Repository) Critical() bool { return true }

func (m *Repository) Ping(context.Context) error { return nil }

func (m *Repository) Create(_ context.Context, w *model.Wishlist) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Conflicts > 0 {
		m.Conflicts--
		return model.ErrConflict
	}
	m.insert(w)
	return nil
}

func (m *Repository) insert(w *model.Wishlist) {
	m.nextID++
	w.ID, w.DateCreated = m.nextID, time.Now()
	for i := range w.Lines {
		m.nextID++
		w.Lines[i].ID, w.Lines[i].WishlistID = m.nextID, w.ID
	}
	m.wishlists[w.ID] = clone(w)
}

func (m *Repository) FirstOrCreate(_ context.Context, w *model.Wishlist) (*model.Wishlist, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var first *model.Wishlist
	for _, c := range m.wishlists {
		if c.OwnerID == w.OwnerID && (first == nil || c.ID < first.ID) {
			first = c
		}
	}
	if first != nil {
		return clone(first), false, nil
	}
	m.insert(w)
	return w, true, nil
}

func (m *Repository) Get(_ context.Context, id int64) (*model.Wishlist, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wishlists[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return clone(w), nil
}

func (m *Repository) GetByKey(_ context.Context, key string) (*model.Wishlist, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.wishlists {
		if w.Key == key {
			return clone(w), nil
		}
	}
	return nil, model.ErrNotFound
}

func (m *Repository) List(_ context.Context, f port.ListFilter) ([]model.Wishlist, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.Wishlist{}
	for _, w := range m.wishlists {
		if f.OwnerID != "" && w.OwnerID != f.OwnerID {
			continue
		}
		if f.PublicOnly && w.Visibility != model.VisibilityPublic {
			continue
		}
		out = append(out, *clone(w))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	total := int64(len(out))
	if f.Offset >= len(out) {
		return []model.Wishlist{}, total, nil
	}
	out = out[f.Offset:]
	if len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out, total, nil
}

func (m *Repository) Update(_ context.Context, w *model.Wishlist) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.wishlists[w.ID]
	if !ok {
		return model.ErrNotFound
	}
	cur.Name, cur.Key, cur.Visibility = w.Name, w.Key, w.Visibility
	return nil
}

func (m *Repository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.wishlists[id]; !ok {
		return model.ErrNotFound
	}
	delete(m.wishlists, id)
	return nil
}

func (m *Repository) AddLine(_ context.Context, l *model.Line) (*model.Line, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.addLine(l)
}

func (m *Repository) addLine(l *model.Line) (*model.Line, error) {
	w, ok := m.wishlists[l.WishlistID]
	if !ok {
		return nil, model.ErrNotFound
	}
	for i := range w.Lines {
		if w.Lines[i].ProductID == l.ProductID {
			w.Lines[i].Quantity = min(w.Lines[i].Quantity+l.Quantity, model.MaxQuantity)
			c := w.Lines[i]
			return &c, nil
		}
	}
	m.nextID++
	c := *l
	c.ID = m.nextID
	w.Lines = append(w.Lines, c)
	return &c, nil
}

func (m *Repository) UpdateLineQuantity(_ context.Context, wishlistID, lineID int64, quantity int) (*model.Line, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wishlists[wishlistID]
	if !ok {
		return nil, model.ErrNotFound
	}
	for i := range w.Lines {
		if w.Lines[i].ID == lineID {
			w.Lines[i].Quantity = quantity
			c := w.Lines[i]
			return &c, nil
		}
	}
	return nil, model.ErrNotFound
}

func (m *Repository) DeleteLine(_ context.Context, wishlistID, lineID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wishlists[wishlistID]
	if !ok {
		return model.ErrNotFound
	}
	for i := range w.Lines {
		if w.Lines[i].ID == lineID {
			w.Lines = append(w.Lines[:i], w.Lines[i+1:]...)
			return nil
		}
	}
	return model.ErrNotFound
}

func (m *Repository) ClearLines(_ context.Context, wishlistID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.wishlists[wishlistID]
	if !ok {
		return 0, nil
	}
	n := int64(len(w.Lines))
	w.Lines = []model.Line{}
	return n, nil
}

func (m *Repository) MoveLine(_ context.Context, fromWishlistID, lineID, toWishlistID int64) (*model.Line, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	src, ok := m.wishlists[fromWishlistID]
	if !ok {
		return nil, model.ErrNotFound
	}
	if _, ok := m.wishlists[toWishlistID]; !ok {
		return nil, model.ErrNotFound
	}
	for i := range src.Lines {
		if src.Lines[i].ID == lineID {
			l := src.Lines[i]
			src.Lines = append(src.Lines[:i], src.Lines[i+1:]...)
			l.WishlistID = toWishlistID
			return m.addLine(&l)
		}
	}
	return nil, model.ErrNotFound
}

func (m *Repository) FindProduct(_ context.Context, ownerID string, productID int64) ([]model.ProductMembership, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.ProductMembership{}
	for _, w := range m.wishlists {
		if w.OwnerID != ownerID {
			continue
		}
		for _, l := range w.Lines {
			if l.ProductID == productID {
				out = append(out, model.ProductMembership{WishlistID: w.ID, WishlistName: w.Name, LineID: l.ID, Quantity: l.Quantity})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WishlistID < out[j].WishlistID })
	return out, nil
}

func (m *Repository) PopularProducts(_ context.Context, limit int) ([]model.ProductStat, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := map[int64]*model.ProductStat{}
	for _, w := range m.wishlists {
		for _, l := range w.Lines {
			s, ok := stats[l.ProductID]
			if !ok {
				s = &model.ProductStat{ProductID: l.ProductID, Title: l.Title}
				stats[l.ProductID] = s
			}
			s.WishlistCount++
			s.TotalQuantity += int64(l.Quantity)
		}
	}
	out := make([]model.ProductStat, 0, len(stats))
	for _, s := range stats {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.WishlistCount != b.WishlistCount {
			return a.WishlistCount > b.WishlistCount
		}
		if a.TotalQuantity != b.TotalQuantity {
			return a.TotalQuantity > b.TotalQuantity
		}
		return a.ProductID < b.ProductID
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

type Users struct {
	mu    sync.Mutex
	users map[string]model.Owner
	Calls int
	Err   error
}

func NewUsers(owners ...model.Owner) *Users {
	u := &Users{users: map[string]model.Owner{}}
	for _, o := range owners {
		u.users[o.ID] = o
	}
	return u
}

func (u *Users) GetUser(_ context.Context, id string) (*model.Owner, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Calls++
	if u.Err != nil {
		return nil, u.Err
	}
	o, ok := u.users[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &o, nil
}
