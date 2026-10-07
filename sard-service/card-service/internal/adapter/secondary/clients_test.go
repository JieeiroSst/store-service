package secondary_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/authservice"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/customerinfo"
	"github.com/JIeeiroSst/card-service/internal/domain"
)

func TestCustomerInfoClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/v1/customers/cus-1":
			w.Write([]byte(`{"id":"cus-1","user_id":7,"full_name":"An","card_eligibility":{"eligible":false,"reasons":["kyc is NONE"]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := customerinfo.NewClient(srv.URL, "tok", time.Second)
	cu, err := c.GetCustomer(context.Background(), "cus-1")
	if err != nil || cu.UserID != 7 || cu.Eligible || len(cu.Reasons) != 1 {
		t.Fatalf("%v %+v", err, cu)
	}
	if _, err := c.GetCustomer(context.Background(), "cus-2"); !domain.IsInvalid(err) {
		t.Fatalf("missing: %v", err)
	}
	if _, err := customerinfo.NewClient(srv.URL, "bad", time.Second).GetCustomer(context.Background(), "cus-1"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("unauthorized: %v", err)
	}
}

func TestAuthServiceClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch r.URL.Path {
		case "/internal/v1/payment-authentications/ok/consume":
			if body["card_id"] != "card-1" || body["amount"].(float64) != 100 || body["currency"] != "VND" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write([]byte(`{"status":"USED"}`))
		case "/internal/v1/payment-authentications/used/consume":
			w.WriteHeader(http.StatusConflict)
		case "/internal/v1/payment-authentications/down/consume":
			w.WriteHeader(http.StatusBadGateway)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := authservice.NewClient(srv.URL, "internal", time.Second)
	ctx := context.Background()
	if err := c.Consume(ctx, "ok", "card-1", 100, "VND"); err != nil {
		t.Fatal(err)
	}
	if err := c.Consume(ctx, "used", "card-1", 100, "VND"); !domain.IsConflict(err) {
		t.Fatalf("used: %v", err)
	}
	if err := c.Consume(ctx, "nope", "card-1", 100, "VND"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if err := c.Consume(ctx, "down", "card-1", 100, "VND"); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("down: %v", err)
	}
	if (authservice.Disabled{}).Enabled() {
		t.Fatal("disabled authenticator")
	}
}
