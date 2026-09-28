package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
)

func newWallet(t *testing.T, h http.HandlerFunc) *Wallet {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewWallet(httpx.New(srv.URL, time.Second, "vending-test"))
}

func TestWalletChargeRetriesWithSameReference(t *testing.T) {
	var calls atomic.Int32
	var refs []string
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/wallets/w-1/withdraw" || r.Header.Get("X-Requested-By") != "vending-test" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Amount      int    `json:"amount"`
			ReferenceID string `json:"reference_id"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		refs = append(refs, body.ReferenceID)
		if calls.Add(1) == 1 {
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
		if body.Amount != 250 {
			t.Errorf("amount %d", body.Amount)
		}
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(map[string]string{"transaction_id": "tx-9", "status": "COMPLETED"})
	})
	id, err := w.Charge(context.Background(), port.ChargeRequest{PaymentID: "pay-1", AmountCents: 250, WalletID: "w-1"})
	if err != nil || id != "tx-9" {
		t.Fatalf("id=%q err=%v", id, err)
	}
	if len(refs) != 2 || refs[0] != "pay-1" || refs[1] != "pay-1" {
		t.Fatalf("references %v", refs)
	}
}

func TestWalletChargeDoesNotRetryDecline(t *testing.T) {
	var calls atomic.Int32
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		rw.WriteHeader(http.StatusUnprocessableEntity)
		rw.Write([]byte(`{"error":"insufficient balance"}`))
	})
	_, err := w.Charge(context.Background(), port.ChargeRequest{PaymentID: "p", AmountCents: 1, WalletID: "w"})
	if err == nil || calls.Load() != 1 || !httpx.IsStatus(err, http.StatusUnprocessableEntity) {
		t.Fatalf("calls=%d err=%v", calls.Load(), err)
	}
}

func TestWalletRefundTreatsAlreadyReversedAsDone(t *testing.T) {
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/transactions/tx-1/reverse":
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{"error":"transaction cannot be reversed"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/transactions/tx-1":
			json.NewEncoder(rw).Encode(map[string]string{"transaction_id": "tx-1", "status": "REVERSED"})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	if err := w.Refund(context.Background(), port.RefundRequest{TransactionID: "tx-1", Reason: "x"}); err != nil {
		t.Fatal(err)
	}
}

func TestWalletRefundReportsRealFailure(t *testing.T) {
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(rw).Encode(map[string]string{"status": "COMPLETED"})
			return
		}
		rw.WriteHeader(http.StatusBadRequest)
	})
	if err := w.Refund(context.Background(), port.RefundRequest{TransactionID: "tx-1"}); err == nil {
		t.Fatal("expected error")
	}
}

type recorder struct{ methods []domain.PaymentMethod }

func (r *recorder) Charge(_ context.Context, req port.ChargeRequest) (string, error) {
	r.methods = append(r.methods, req.Method)
	return "ok", nil
}

func (r *recorder) Refund(context.Context, port.RefundRequest) error { return nil }

func (r *recorder) Lookup(context.Context, port.LookupRequest) (string, bool, error) {
	return "", false, nil
}

func TestRouter(t *testing.T) {
	wallet, terminal := &recorder{}, &recorder{}
	r := NewRouter(wallet, terminal)
	r.Charge(context.Background(), port.ChargeRequest{Method: domain.PaymentWallet})
	r.Charge(context.Background(), port.ChargeRequest{Method: domain.PaymentCard})
	if len(wallet.methods) != 1 || len(terminal.methods) != 1 {
		t.Fatalf("wallet=%v terminal=%v", wallet.methods, terminal.methods)
	}
	if _, err := NewRouter(nil, terminal).Charge(context.Background(), port.ChargeRequest{Method: domain.PaymentWallet}); err == nil {
		t.Fatal("wallet payments must fail when the wallet service is not configured")
	}
}

func TestWalletChargeTimeoutFoundByLookup(t *testing.T) {
	var withdraws atomic.Int32
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			withdraws.Add(1)
			rw.WriteHeader(http.StatusGatewayTimeout)
		case r.URL.Path == "/api/v1/wallets/w-1/transactions/by-reference/pay-1":
			json.NewEncoder(rw).Encode(map[string]string{"transaction_id": "tx-landed", "status": "COMPLETED"})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	id, err := w.Charge(context.Background(), port.ChargeRequest{PaymentID: "pay-1", AmountCents: 100, WalletID: "w-1"})
	if err != nil || id != "tx-landed" || withdraws.Load() != walletAttempts {
		t.Fatalf("id=%q err=%v withdraws=%d", id, err, withdraws.Load())
	}
}

func TestWalletChargeTimeoutNotFoundIsPending(t *testing.T) {
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
		rw.WriteHeader(http.StatusNotFound)
	})
	_, err := w.Charge(context.Background(), port.ChargeRequest{PaymentID: "pay-1", AmountCents: 100, WalletID: "w-1"})
	if !errors.Is(err, domain.ErrPaymentPending) {
		t.Fatalf("an unanswered charge must be reported as pending, got %v", err)
	}
}

func TestWalletDeclineIsNotPending(t *testing.T) {
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			t.Error("a clear decline must not trigger a lookup")
		}
		rw.WriteHeader(http.StatusUnprocessableEntity)
	})
	_, err := w.Charge(context.Background(), port.ChargeRequest{PaymentID: "pay-1", AmountCents: 100, WalletID: "w-1"})
	if err == nil || errors.Is(err, domain.ErrPaymentPending) {
		t.Fatalf("got %v", err)
	}
}

func TestWalletLookup(t *testing.T) {
	status := http.StatusNotFound
	w := newWallet(t, func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(status)
		if status == http.StatusOK {
			json.NewEncoder(rw).Encode(map[string]string{"transaction_id": "tx-1"})
		}
	})
	req := port.LookupRequest{PaymentID: "p", WalletID: "w"}
	if _, found, err := w.Lookup(context.Background(), req); found || err != nil {
		t.Fatalf("404: found=%v err=%v", found, err)
	}
	status = http.StatusOK
	if id, found, err := w.Lookup(context.Background(), req); !found || err != nil || id != "tx-1" {
		t.Fatalf("200: id=%q found=%v err=%v", id, found, err)
	}
	status = http.StatusServiceUnavailable
	if _, found, err := w.Lookup(context.Background(), req); found || err == nil {
		t.Fatalf("503 must be an error, not 'not found': found=%v err=%v", found, err)
	}
}
