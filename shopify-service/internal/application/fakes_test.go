package application

import (
	"context"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

type fakeProductRepo struct {
	byShopifyID map[string]*model.Product
	nextID      int64
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{byShopifyID: map[string]*model.Product{}}
}

func (r *fakeProductRepo) Upsert(_ context.Context, p *model.Product) (*model.Product, error) {
	if existing, ok := r.byShopifyID[p.ShopifyID]; ok {
		p.ID = existing.ID
	} else {
		r.nextID++
		p.ID = r.nextID
	}
	clone := *p
	r.byShopifyID[p.ShopifyID] = &clone
	return p, nil
}

func (r *fakeProductRepo) GetByID(_ context.Context, id int64) (*model.Product, error) {
	for _, p := range r.byShopifyID {
		if p.ID == id {
			clone := *p
			return &clone, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r *fakeProductRepo) List(_ context.Context, limit, offset int) ([]model.Product, int64, error) {
	var out []model.Product
	for _, p := range r.byShopifyID {
		out = append(out, *p)
	}
	return out, int64(len(out)), nil
}

func (r *fakeProductRepo) DeleteByShopifyID(_ context.Context, shopifyID string) error {
	delete(r.byShopifyID, shopifyID)
	return nil
}

type fakeOrderRepo struct {
	byShopifyID map[string]*model.Order
	nextID      int64
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{byShopifyID: map[string]*model.Order{}}
}

func (r *fakeOrderRepo) Upsert(_ context.Context, o *model.Order) (*model.Order, error) {
	if existing, ok := r.byShopifyID[o.ShopifyID]; ok {
		o.ID = existing.ID
	} else {
		r.nextID++
		o.ID = r.nextID
	}
	clone := *o
	r.byShopifyID[o.ShopifyID] = &clone
	return o, nil
}

func (r *fakeOrderRepo) GetByID(_ context.Context, id int64) (*model.Order, error) {
	for _, o := range r.byShopifyID {
		if o.ID == id {
			clone := *o
			return &clone, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r *fakeOrderRepo) List(_ context.Context, limit, offset int) ([]model.Order, int64, error) {
	var out []model.Order
	for _, o := range r.byShopifyID {
		out = append(out, *o)
	}
	return out, int64(len(out)), nil
}

type fakeEventRepo struct {
	seen map[string]bool
}

func newFakeEventRepo() *fakeEventRepo { return &fakeEventRepo{seen: map[string]bool{}} }

func (r *fakeEventRepo) Exists(_ context.Context, id string) (bool, error) { return r.seen[id], nil }

func (r *fakeEventRepo) Create(_ context.Context, e *model.WebhookEvent) error {
	r.seen[e.WebhookID] = true
	return nil
}

type fakeShopify struct {
	productPages []port.Page[model.Product]
	orderPages   []port.Page[model.Order]
	products     map[string]*model.Product
	orders       map[string]*model.Order
	created      *port.CreateProductInput
	getCalls     int
}

func newFakeShopify() *fakeShopify {
	return &fakeShopify{products: map[string]*model.Product{}, orders: map[string]*model.Order{}}
}

func (s *fakeShopify) ListProducts(_ context.Context, after string) (*port.Page[model.Product], error) {
	return pageAt(s.productPages, after), nil
}

func (s *fakeShopify) GetProduct(_ context.Context, id string) (*model.Product, error) {
	s.getCalls++
	p, ok := s.products[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	clone := *p
	return &clone, nil
}

func (s *fakeShopify) CreateProduct(_ context.Context, in port.CreateProductInput) (*model.Product, error) {
	s.created = &in
	return &model.Product{ShopifyID: "gid://shopify/Product/99", Title: in.Title, Status: in.Status}, nil
}

func (s *fakeShopify) ListOrders(_ context.Context, after string) (*port.Page[model.Order], error) {
	return pageAt(s.orderPages, after), nil
}

func (s *fakeShopify) GetOrder(_ context.Context, id string) (*model.Order, error) {
	s.getCalls++
	o, ok := s.orders[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	clone := *o
	return &clone, nil
}

// pageAt treats the cursor as the page index ("" is page 0, "1" page 1, ...).
func pageAt[T any](pages []port.Page[T], after string) *port.Page[T] {
	idx := 0
	if after != "" {
		idx = int(after[0] - '0')
	}
	p := pages[idx]
	return &p
}

type fakeVerifier struct{ ok bool }

func (v fakeVerifier) Verify([]byte, string) bool { return v.ok }
