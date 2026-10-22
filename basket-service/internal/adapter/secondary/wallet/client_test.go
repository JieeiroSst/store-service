package wallet

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

func TestPaySendsTheShopMerchantAndIdempotencyKey(t *testing.T) {
	var body map[string]any
	var header string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("Idempotency-Key")
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["pin"] != "123456" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"error":"incorrect PIN"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"tx-9","status":"COMPLETED"}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "66666666-6666-6666-6666-666666666666", 0)

	tx, err := c.Pay(context.Background(), port.WalletPayment{UserID: 7, WalletID: "w-1", Amount: 300000, Pin: "123456", IdempotencyKey: "keiki-basket-1"})
	if err != nil || tx != "tx-9" {
		t.Fatalf("pay: %q %v", tx, err)
	}
	if body["user_id"] != "7" || body["merchant_id"] != "66666666-6666-6666-6666-666666666666" || body["amount"] != "300000.00" || header != "keiki-basket-1" {
		t.Fatalf("unexpected request: %v header=%q", body, header)
	}

	_, err = c.Pay(context.Background(), port.WalletPayment{UserID: 7, WalletID: "w-1", Amount: 1, Pin: "000000", IdempotencyKey: "k"})
	if !errors.Is(err, common.ErrPaymentFailed) || err.Error() != "payment failed: incorrect PIN" {
		t.Fatalf("declined payment: %v", err)
	}
}
