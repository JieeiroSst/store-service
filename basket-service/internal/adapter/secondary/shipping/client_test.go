package shipping

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

var to = port.ShippingAddress{Name: "Quân", Phone: "0912345678", Street: "24 Lê Văn Hưu", WardCode: "20109", DistrictID: 1442, ProvinceID: 202}

func TestQuotePicksTheCheapestRoute(t *testing.T) {
	var got map[string]any
	var key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/shipments/quote" {
			t.Errorf("path %s", r.URL.Path)
		}
		key = r.Header.Get("X-API-Key")
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"options":[{"fee":41000,"service_name":"Nhanh"},{"fee":28000,"service_name":"Chuẩn","warehouse_id":2,"service_id":53320}],"recommended":{"CHEAPEST":{"fee":28000,"service_name":"Chuẩn","warehouse_id":2,"service_id":53320}}}`))
	}))
	defer srv.Close()

	q, err := New(srv.URL, "basket-key", 0).Quote(context.Background(), to, []port.ShippingItem{{Name: "Áo", Quantity: 3, Price: 150000}}, 9_000_000)
	if err != nil || q.Fee != 28000 || q.ServiceName != "Chuẩn" || q.WarehouseID != 2 || q.ServiceID != 53320 {
		t.Fatalf("q=%+v err=%v", q, err)
	}
	parcel := got["parcel"].(map[string]any)
	if key != "basket-key" || parcel["weight_gram"] != float64(1500) || parcel["height_cm"] != float64(25) || got["insurance_value"] != float64(5_000_000) {
		t.Fatalf("key=%q body=%v", key, got)
	}
	if got["recipient"].(map[string]any)["ward_code"] != "20109" {
		t.Fatalf("recipient %v", got["recipient"])
	}
}

func TestQuoteErrors(t *testing.T) {
	for status, want := range map[int]error{http.StatusUnprocessableEntity: common.ErrInvalidRequest, http.StatusBadRequest: common.ErrInvalidRequest, http.StatusBadGateway: common.ErrUpstream, http.StatusUnauthorized: common.ErrUpstream} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"no route"}`))
		}))
		_, err := New(srv.URL, "k", 0).Quote(context.Background(), to, nil, 0)
		srv.Close()
		if !errors.Is(err, want) {
			t.Fatalf("%d: %v", status, err)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"options":[],"recommended":{}}`))
	}))
	defer srv.Close()
	if _, err := New(srv.URL, "k", 0).Quote(context.Background(), to, nil, 0); !errors.Is(err, common.ErrInvalidRequest) {
		t.Fatalf("no route: %v", err)
	}
}
