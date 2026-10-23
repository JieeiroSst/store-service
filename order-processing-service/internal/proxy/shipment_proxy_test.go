package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

func TestShipmentProxyCreate(t *testing.T) {
	var got entity.ShipmentRequest
	var key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/shipments" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		key = r.Header.Get("X-API-Key")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":9,"code":"SHP-9","status":"CREATED","carrier_order_code":"GHN123","quoted_fee":32000}`))
	}))
	defer srv.Close()

	s, err := NewShipmentProxy(srv.URL+"/", "order-key").CreateShipment(context.Background(), entity.ShipmentRequest{ClientOrderCode: "KS00000012"})
	if err != nil || s.ID != 9 || s.CarrierOrderCode != "GHN123" || s.QuotedFee != 32000 {
		t.Fatalf("shipment=%+v err=%v", s, err)
	}
	if key != "order-key" || got.ClientOrderCode != "KS00000012" {
		t.Fatalf("key=%q body=%+v", key, got)
	}
}

func TestShipmentProxyErrors(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
		ship   bool
	}{
		{http.StatusUnprocessableEntity, `{"error":"no route","shipment":{"id":3,"code":"SHP-3","status":"REJECTED"}}`, ErrShipmentRejected, true},
		{http.StatusBadRequest, `{"error":"recipient: phone must be a Vietnamese number"}`, ErrShipmentInvalid, false},
		{http.StatusUnauthorized, `{"error":"missing or invalid X-API-Key"}`, ErrShipmentAuth, false},
		{http.StatusBadGateway, `{"error":"ghn down"}`, nil, false},
	}
	for _, c := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(c.status)
			_, _ = w.Write([]byte(c.body))
		}))
		s, err := NewShipmentProxy(srv.URL, "k").CreateShipment(context.Background(), entity.ShipmentRequest{})
		srv.Close()
		if err == nil {
			t.Fatalf("%d: expected an error", c.status)
		}
		if c.want != nil && !errors.Is(err, c.want) {
			t.Fatalf("%d: err=%v", c.status, err)
		}
		if c.want == nil && (errors.Is(err, ErrShipmentRejected) || errors.Is(err, ErrShipmentInvalid) || errors.Is(err, ErrShipmentAuth)) {
			t.Fatalf("%d must stay retryable: %v", c.status, err)
		}
		if c.ship != (s != nil) {
			t.Fatalf("%d: shipment=%+v", c.status, s)
		}
	}
}
