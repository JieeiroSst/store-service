package catalogue

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/basket-service/common"
)

func TestClientMapsCatalogueResponses(t *testing.T) {
	var allocated []int
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/stockrecords/5", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 5, "product_id": 9, "price_currency": "VND", "price": 120000, "num_available": 2})
	})
	mux.HandleFunc("POST /v1/stockrecords/5/allocate", func(w http.ResponseWriter, r *http.Request) {
		var b struct{ Quantity int }
		_ = json.NewDecoder(r.Body).Decode(&b)
		if b.Quantity > 2 {
			w.WriteHeader(http.StatusConflict)
			return
		}
		allocated = append(allocated, b.Quantity)
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c := New(srv.URL, 0)

	r, err := c.StockRecord(context.Background(), 5)
	if err != nil || r.Price != 120000 || r.ProductID != 9 || *r.NumAvailable != 2 {
		t.Fatalf("unexpected record: %+v %v", r, err)
	}
	if _, err := c.StockRecord(context.Background(), 6); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("missing record: %v", err)
	}
	if err := c.Allocate(context.Background(), 5, 2); err != nil || len(allocated) != 1 {
		t.Fatalf("allocate: %v", err)
	}
	if err := c.Allocate(context.Background(), 5, 3); !errors.Is(err, common.ErrOutOfStock) {
		t.Fatalf("over-allocate: %v", err)
	}
}
