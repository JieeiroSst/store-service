package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
)

type stubOptions struct {
	port.OptionService
	create func(model.Option) (*model.Option, error)
}

func (s stubOptions) Create(_ context.Context, o model.Option) (*model.Option, error) {
	return s.create(o)
}
func (s stubOptions) Get(context.Context, int64) (*model.Option, error) {
	return nil, model.ErrNotFound
}

func do(t *testing.T, svc port.OptionService, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	NewOptionHandler(svc).Register(mux)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestStatusMapping(t *testing.T) {
	var next error
	svc := stubOptions{create: func(o model.Option) (*model.Option, error) {
		if next != nil {
			return nil, next
		}
		o.ID = 7
		return &o, nil
	}}

	for _, tc := range []struct {
		name   string
		err    error
		body   string
		status int
	}{
		{"created", nil, `{"name":"Gift"}`, http.StatusCreated},
		{"invalid", model.Invalid("nope"), `{"name":"Gift"}`, http.StatusBadRequest},
		{"conflict", model.ErrConflict, `{"name":"Gift"}`, http.StatusConflict},
		{"internal", context.DeadlineExceeded, `{"name":"Gift"}`, http.StatusInternalServerError},
		{"unknown field", nil, `{"nam":"Gift"}`, http.StatusBadRequest},
		{"malformed", nil, `{`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next = tc.err
			if got := do(t, svc, "POST", "/v1/options", tc.body).Code; got != tc.status {
				t.Errorf("status = %d, want %d", got, tc.status)
			}
		})
	}
}

func TestPathID(t *testing.T) {
	svc := stubOptions{}
	if got := do(t, svc, "GET", "/v1/options/abc", "").Code; got != http.StatusBadRequest {
		t.Errorf("non-numeric id: status %d", got)
	}
	if got := do(t, svc, "GET", "/v1/options/5", "").Code; got != http.StatusNotFound {
		t.Errorf("missing option: status %d", got)
	}
}

func TestProductJSONSummarisesStock(t *testing.T) {
	five, two, zero := 5, 2, 0
	p := model.Product{ID: 1, Title: "Áo", StockRecords: []model.StockRecord{
		{ID: 1, Price: 300, PriceCurrency: "VND", NumInStock: &five, NumAllocated: 5},
		{ID: 2, Price: 250, PriceCurrency: "VND", NumInStock: &two},
		{ID: 3, Price: 100, PriceCurrency: "VND", NumInStock: &zero},
	}}
	out := toProductJSON(p)
	if out.Price == nil || *out.Price != 250 || out.PriceCurrency != "VND" || !out.InStock {
		t.Fatalf("cheapest purchasable offer must win: %+v", out)
	}
	if out.NumAvailable == nil || *out.NumAvailable != 2 {
		t.Fatalf("available = %v, want 2", out.NumAvailable)
	}

	empty := toProductJSON(model.Product{ID: 2})
	if empty.Price != nil || empty.InStock || empty.NumAvailable != nil || empty.StockRecords == nil {
		t.Fatalf("product without stock must be unpriced: %+v", empty)
	}

	untracked := toProductJSON(model.Product{ID: 3, StockRecords: []model.StockRecord{{Price: 9}}})
	if untracked.NumAvailable != nil || !untracked.InStock {
		t.Fatalf("untracked stock is always available: %+v", untracked)
	}
}
