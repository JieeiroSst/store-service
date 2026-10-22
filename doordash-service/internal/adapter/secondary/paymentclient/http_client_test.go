package paymentclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

func newWallet(t *testing.T, owner string) (*httptest.Server, *[]map[string]any, *[]string) {
	t.Helper()
	var authorized []map[string]any
	var settled []string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/cards/card-1", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "card-1", "wallet_id": "wallet-1"})
	})
	mux.HandleFunc("GET /api/v1/users/{user}/wallets", func(w http.ResponseWriter, r *http.Request) {
		items := []map[string]any{}
		if r.PathValue("user") == owner {
			items = append(items, map[string]any{"id": "wallet-1"})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})
	mux.HandleFunc("POST /api/v1/transactions/authorize", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		authorized = append(authorized, body)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "tx-1", "status": "AUTHORIZED"})
	})
	mux.HandleFunc("POST /api/v1/transactions/{id}/{action}", func(w http.ResponseWriter, r *http.Request) {
		settled = append(settled, r.PathValue("action")+":"+r.PathValue("id"))
		_ = json.NewEncoder(w).Encode(map[string]any{"id": r.PathValue("id")})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &authorized, &settled
}

func client(url string) port.PaymentClient {
	cfg := &config.Config{
		PaymentService: config.ExternalServiceConfig{BaseURL: url, Timeout: "2s"},
		Wallet:         config.WalletConfig{MerchantID: "merchant-1", Currency: "VND"},
	}
	return NewPaymentClient(cfg)
}

func TestAuthorizeChargesTheCustomersOwnCard(t *testing.T) {
	srv, authorized, _ := newWallet(t, "7")
	auth, err := client(srv.URL).AuthorizePayment(context.Background(), port.AuthorizePaymentInput{CustomerID: "7", Amount: 125000, PaymentMethodID: "card-1"})
	if err != nil {
		t.Fatal(err)
	}
	if auth.Status != model.PaymentStatusAuthorized || auth.PaymentID != "tx-1" {
		t.Fatalf("unexpected authorization: %+v", auth)
	}
	body := (*authorized)[0]
	if body["card_id"] != "card-1" || body["merchant_id"] != "merchant-1" || body["amount"] != "125000.00" || body["currency"] != "VND" {
		t.Fatalf("unexpected authorize request: %v", body)
	}
	if key, _ := body["idempotency_key"].(string); len(key) < 20 {
		t.Fatalf("idempotency key too weak: %q", key)
	}
}

func TestAuthorizeRejectsSomeoneElsesCard(t *testing.T) {
	srv, authorized, _ := newWallet(t, "7")
	auth, err := client(srv.URL).AuthorizePayment(context.Background(), port.AuthorizePaymentInput{CustomerID: "8", Amount: 10, PaymentMethodID: "card-1"})
	if err != nil {
		t.Fatal(err)
	}
	if auth.Status != model.PaymentStatusFailed || len(*authorized) != 0 {
		t.Fatalf("foreign card must not be charged: %+v %v", auth, *authorized)
	}
	auth, _ = client(srv.URL).AuthorizePayment(context.Background(), port.AuthorizePaymentInput{CustomerID: "7", Amount: 10, PaymentMethodID: "missing"})
	if auth.Status != model.PaymentStatusFailed {
		t.Fatalf("unknown card must fail, got %+v", auth)
	}
}

func TestCaptureAndVoid(t *testing.T) {
	srv, _, settled := newWallet(t, "7")
	c := client(srv.URL)
	if err := c.CapturePayment(context.Background(), "tx-1"); err != nil {
		t.Fatal(err)
	}
	if err := c.VoidPayment(context.Background(), "tx-2"); err != nil {
		t.Fatal(err)
	}
	if len(*settled) != 2 || (*settled)[0] != "capture:tx-1" || (*settled)[1] != "void:tx-2" {
		t.Fatalf("unexpected settlements: %v", *settled)
	}
}
