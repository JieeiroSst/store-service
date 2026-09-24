package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

func newClient(t *testing.T, h http.HandlerFunc, minor int64) *Client {
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(&config.Config{Wallet: config.WalletConfig{
		UpstreamConfig:   config.UpstreamConfig{BaseURL: srv.URL, Timeout: 2 * time.Second},
		HospitalWalletID: "hospital-wallet", MinorUnit: minor, UserPrefix: "patient-",
	}})
}

func TestChargeMovesMinorUnitsToTheHospitalWallet(t *testing.T) {
	var transfer map[string]any
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/wallets/user/patient-5":
			_, _ = w.Write([]byte(`{"wallet_id":"w-5"}`))
		case "POST /api/v1/transfers":
			_ = json.NewDecoder(r.Body).Decode(&transfer)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"transfer_id":"t-1"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}, 100)

	id, err := c.Charge(context.Background(), port.ChargeInput{PatientID: 5, Amount: 60.55, Reference: "ref-1", Description: "bill"})
	if err != nil || id != "t-1" {
		t.Fatalf("id=%q err=%v", id, err)
	}
	if transfer["sender_wallet_id"] != "w-5" || transfer["receiver_wallet_id"] != "hospital-wallet" ||
		transfer["amount"] != float64(6055) || transfer["reference_id"] != "ref-1" {
		t.Errorf("transfer = %v", transfer)
	}
}

func TestChargeRespectsTheCurrencyMinorUnit(t *testing.T) {
	var transfer map[string]any
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"wallet_id":"w"}`))
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&transfer)
		_, _ = w.Write([]byte(`{"transfer_id":"t"}`))
	}, 1) // e.g. VND has no minor unit

	if _, err := c.Charge(context.Background(), port.ChargeInput{PatientID: 1, Amount: 250000}); err != nil {
		t.Fatal(err)
	}
	if transfer["amount"] != float64(250000) {
		t.Errorf("amount = %v", transfer["amount"])
	}
}

func TestChargeFailuresAreClassified(t *testing.T) {
	cases := map[string]struct {
		walletStatus, transferStatus int
		wantPayment, wantUpstream    bool
	}{
		"no wallet":          {walletStatus: 404, wantPayment: true},
		"insufficient funds": {walletStatus: 200, transferStatus: 422, wantPayment: true},
		"wallet frozen":      {walletStatus: 200, transferStatus: 400, wantPayment: true},
		"wallet-service 500": {walletStatus: 200, transferStatus: 500, wantUpstream: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(tc.walletStatus)
					_, _ = w.Write([]byte(`{"wallet_id":"w"}`))
					return
				}
				w.WriteHeader(tc.transferStatus)
				_, _ = w.Write([]byte(`{"error":"nope"}`))
			}, 100)
			_, err := c.Charge(context.Background(), port.ChargeInput{PatientID: 1, Amount: 1})
			if errors.Is(err, model.ErrPaymentFailed) != tc.wantPayment || errors.Is(err, model.ErrUpstream) != tc.wantUpstream {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestRefundReversesTheTransfer(t *testing.T) {
	var path string
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) { path = r.Method + " " + r.URL.Path }, 100)
	if err := c.Refund(context.Background(), "t-9", "reopened"); err != nil {
		t.Fatal(err)
	}
	if path != "POST /api/v1/transfers/t-9/reverse" {
		t.Errorf("request = %s", path)
	}
}

func TestEnabledNeedsBothURLAndHospitalWallet(t *testing.T) {
	if New(&config.Config{}).Enabled() {
		t.Error("unconfigured client must be disabled")
	}
	if New(&config.Config{Wallet: config.WalletConfig{UpstreamConfig: config.UpstreamConfig{BaseURL: "http://x"}}}).Enabled() {
		t.Error("client without a hospital wallet must be disabled")
	}
}

func TestRefundOfAnAlreadyReversedTransferSucceeds(t *testing.T) {
	reverses := 0
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"transfer_id":"t-9","status":"REVERSED"}`))
		case http.MethodPost:
			reverses++
			http.Error(w, `{"error":"transfer cannot be reversed"}`, http.StatusBadRequest)
		}
	}, 100)
	if err := c.Refund(context.Background(), "t-9", "retry"); err != nil || reverses != 0 {
		t.Errorf("err=%v reverses=%d", err, reverses)
	}
}

func TestChargeUsesTheLinkedUserIDWhenGiven(t *testing.T) {
	var path string
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			path = r.URL.Path
			_, _ = w.Write([]byte(`{"wallet_id":"w"}`))
			return
		}
		_, _ = w.Write([]byte(`{"transfer_id":"t"}`))
	}, 100)
	if _, err := c.Charge(context.Background(), port.ChargeInput{PatientID: 5, UserID: "user-77", Amount: 1}); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/wallets/user/user-77" {
		t.Errorf("looked up %s", path)
	}
}
