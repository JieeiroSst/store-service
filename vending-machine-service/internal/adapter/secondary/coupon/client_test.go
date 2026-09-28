package coupon

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

func newClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewClient(httpx.New(srv.URL, time.Second, ""))
}

func TestQuoteConvertsUnits(t *testing.T) {
	c := newClient(t, func(rw http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/api/v1/coupons/validate" || body["purchase_amount"] != 1.5 || body["code"] != "SAVE" {
			t.Errorf("request %s %v", r.URL.Path, body)
		}
		json.NewEncoder(rw).Encode(map[string]any{"discount": 0.3})
	})
	d, err := c.Quote(context.Background(), "SAVE", 150)
	if err != nil || d != 30 {
		t.Fatalf("discount=%d err=%v", d, err)
	}
}

func TestRedeemSendsOrderNumber(t *testing.T) {
	c := newClient(t, func(rw http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/api/v1/coupons/apply" || body["order_id"] != float64(42) {
			t.Errorf("request %s %v", r.URL.Path, body)
		}
		json.NewEncoder(rw).Encode(map[string]any{"id": 1})
	})
	if err := c.Redeem(context.Background(), "SAVE", 42, 150); err != nil {
		t.Fatal(err)
	}
}

func TestErrorMapping(t *testing.T) {
	status := http.StatusConflict
	c := newClient(t, func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(status)
		rw.Write([]byte(`{"error":"coupon expired"}`))
	})
	_, err := c.Quote(context.Background(), "OLD", 100)
	if !errors.Is(err, domain.ErrCouponRejected) || err.Error() != "coupon rejected: coupon expired" {
		t.Fatalf("got %v", err)
	}
	status = http.StatusInternalServerError
	if _, err := c.Quote(context.Background(), "OLD", 100); !errors.Is(err, domain.ErrUpstream) {
		t.Fatalf("got %v", err)
	}
	if _, err := (Disabled{}).Quote(context.Background(), "X", 1); !errors.Is(err, domain.ErrCouponRejected) {
		t.Fatalf("got %v", err)
	}
}
