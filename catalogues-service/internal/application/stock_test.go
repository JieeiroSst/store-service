package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type fakeProducts struct {
	port.ProductRepository
	known map[int64]bool
}

func (f fakeProducts) Get(_ context.Context, id int64) (*model.Product, error) {
	if !f.known[id] {
		return nil, model.ErrNotFound
	}
	return &model.Product{ID: id}, nil
}

func (f fakeProducts) List(context.Context, model.ProductFilter) ([]model.Product, error) {
	return []model.Product{{ID: 1}, {ID: 2}}, nil
}

type fakeStock struct {
	byID map[int64]*model.StockRecord
	next int64
}

func newFakeStock() *fakeStock { return &fakeStock{byID: map[int64]*model.StockRecord{}} }

func (f *fakeStock) Create(_ context.Context, s *model.StockRecord) error {
	f.next++
	s.ID = f.next
	cp := *s
	f.byID[s.ID] = &cp
	return nil
}

func (f *fakeStock) Get(_ context.Context, id int64) (*model.StockRecord, error) {
	s, ok := f.byID[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeStock) ListByProducts(_ context.Context, ids []int64) ([]model.StockRecord, error) {
	var out []model.StockRecord
	for _, id := range ids {
		for _, s := range f.byID {
			if s.ProductID == id {
				out = append(out, *s)
			}
		}
	}
	return out, nil
}

func (f *fakeStock) Update(_ context.Context, s *model.StockRecord) error {
	cp := *s
	f.byID[s.ID] = &cp
	return nil
}

func (f *fakeStock) Delete(_ context.Context, id int64) error {
	delete(f.byID, id)
	return nil
}

func (f *fakeStock) Allocate(_ context.Context, id int64, qty int) error {
	s, ok := f.byID[id]
	if !ok {
		return model.ErrNotFound
	}
	if !s.CanSupply(qty) {
		return model.ErrOutOfStock
	}
	s.NumAllocated += qty
	return nil
}

func (f *fakeStock) Release(_ context.Context, id int64, qty int) error {
	s := f.byID[id]
	s.NumAllocated = max(s.NumAllocated-qty, 0)
	return nil
}

func intp(v int) *int { return &v }

func TestStockRecordCreateValidates(t *testing.T) {
	svc := NewStockRecordService(newFakeStock(), fakeProducts{known: map[int64]bool{1: true}})
	ctx := context.Background()

	s, err := svc.Create(ctx, model.StockRecord{ProductID: 1, Price: 99000, PriceCurrency: " vnd ", NumInStock: intp(5), NumAllocated: 3})
	if err != nil {
		t.Fatal(err)
	}
	if s.PriceCurrency != "VND" || s.NumAllocated != 0 {
		t.Fatalf("currency must be normalised and allocations reset: %+v", s)
	}
	for name, bad := range map[string]model.StockRecord{
		"negative price": {ProductID: 1, Price: -1},
		"bad currency":   {ProductID: 1, Price: 1, PriceCurrency: "DONG"},
		"negative stock": {ProductID: 1, Price: 1, NumInStock: intp(-2)},
	} {
		if _, err := svc.Create(ctx, bad); !errors.Is(err, model.ErrInvalid) {
			t.Errorf("%s: got %v, want ErrInvalid", name, err)
		}
	}
	if _, err := svc.Create(ctx, model.StockRecord{ProductID: 404, Price: 1}); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("unknown product: got %v", err)
	}
}

func TestStockAllocationNeverOversells(t *testing.T) {
	stock := newFakeStock()
	svc := NewStockRecordService(stock, fakeProducts{known: map[int64]bool{1: true}})
	ctx := context.Background()
	s, _ := svc.Create(ctx, model.StockRecord{ProductID: 1, Price: 10, NumInStock: intp(3)})

	if got, err := svc.Allocate(ctx, s.ID, 2); err != nil || *got.Available() != 1 {
		t.Fatalf("allocate 2: %v %+v", err, got)
	}
	if _, err := svc.Allocate(ctx, s.ID, 2); !errors.Is(err, model.ErrOutOfStock) {
		t.Fatalf("over-allocation must fail, got %v", err)
	}
	if _, err := svc.Allocate(ctx, s.ID, 0); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("zero quantity must be invalid, got %v", err)
	}
	if got, _ := svc.Release(ctx, s.ID, 5); got.NumAllocated != 0 {
		t.Fatalf("release must not go negative: %+v", got)
	}

	untracked, _ := svc.Create(ctx, model.StockRecord{ProductID: 1, Price: 10})
	if _, err := svc.Allocate(ctx, untracked.ID, 1000); err != nil {
		t.Fatalf("untracked stock is unlimited: %v", err)
	}
}

func TestProductsCarryTheirStockRecords(t *testing.T) {
	stock := newFakeStock()
	_ = stock.Create(context.Background(), &model.StockRecord{ProductID: 2, Price: 5, PriceCurrency: "VND"})
	svc := NewProductService(fakeProducts{known: map[int64]bool{1: true, 2: true}}, stock)

	ps, err := svc.List(context.Background(), model.ProductFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps[0].StockRecords) != 0 || len(ps[1].StockRecords) != 1 {
		t.Fatalf("stock records not attached: %+v", ps)
	}
	p, _ := svc.Get(context.Background(), 2)
	if len(p.StockRecords) != 1 || p.StockRecords[0].Price != 5 {
		t.Fatalf("get must attach stock: %+v", p)
	}
}
