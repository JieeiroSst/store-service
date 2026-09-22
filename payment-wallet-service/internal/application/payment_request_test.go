package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

func newPaymentRequestTestDeps() (*fakeState, port.PaymentRequestUsecase) {
	s := newFakeState()
	wallets := &fakeWalletRepo{s: s}
	transactions := NewTransactionService(wallets, &fakeTransactionRepo{s: s}, &fakeTransferRepo{s: s})
	return s, NewPaymentRequestService(wallets, &fakePaymentRequestRepo{s: s}, transactions)
}

func TestPaymentRequest(t *testing.T) {
	t.Run("QR-style request: anyone can pay it", func(t *testing.T) {
		s, svc := newPaymentRequestTestDeps()
		s.wallets["merchant"] = &model.Wallet{WalletID: "merchant", UserID: "m1", Currency: "USD", Status: model.WalletActive}
		s.wallets["payer"] = &model.Wallet{WalletID: "payer", UserID: "u1", Currency: "USD", Status: model.WalletActive, Balance: 1000}

		req, err := svc.CreatePaymentRequest(context.Background(), "merchant", nil, 250, "coffee", 0)
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		transfer, err := svc.Pay(context.Background(), req.PaymentRequestID, "payer")
		if err != nil {
			t.Fatalf("pay: %v", err)
		}
		if transfer.Amount != 250 {
			t.Errorf("amount = %d, want 250", transfer.Amount)
		}
		if got := s.wallets["merchant"].Balance; got != 250 {
			t.Errorf("merchant balance = %d, want 250", got)
		}
		if s.paymentRequests[req.PaymentRequestID].Status != model.PaymentRequestPaid {
			t.Error("request should be marked PAID")
		}
	})

	t.Run("request money: only the designated payer can pay", func(t *testing.T) {
		s, svc := newPaymentRequestTestDeps()
		s.wallets["requester"] = &model.Wallet{WalletID: "requester", UserID: "u1", Currency: "USD", Status: model.WalletActive}
		s.wallets["designated"] = &model.Wallet{WalletID: "designated", UserID: "u2", Currency: "USD", Status: model.WalletActive, Balance: 1000}
		s.wallets["stranger"] = &model.Wallet{WalletID: "stranger", UserID: "u3", Currency: "USD", Status: model.WalletActive, Balance: 1000}

		payerID := "designated"
		req, err := svc.CreatePaymentRequest(context.Background(), "requester", &payerID, 100, "split dinner", 0)
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		if _, err := svc.Pay(context.Background(), req.PaymentRequestID, "stranger"); !errors.Is(err, port.ErrPaymentRequestPayerMismatch) {
			t.Fatalf("err = %v, want ErrPaymentRequestPayerMismatch", err)
		}
		if _, err := svc.Pay(context.Background(), req.PaymentRequestID, "designated"); err != nil {
			t.Fatalf("designated payer should be able to pay: %v", err)
		}
	})

	t.Run("rejects paying an expired request", func(t *testing.T) {
		s, svc := newPaymentRequestTestDeps()
		s.wallets["requester"] = &model.Wallet{WalletID: "requester", UserID: "u1", Currency: "USD", Status: model.WalletActive}
		s.wallets["payer"] = &model.Wallet{WalletID: "payer", UserID: "u2", Currency: "USD", Status: model.WalletActive, Balance: 1000}

		req, _ := svc.CreatePaymentRequest(context.Background(), "requester", nil, 100, "", time.Minute)
		past := time.Now().Add(-time.Hour)
		s.paymentRequests[req.PaymentRequestID].ExpiresAt = &past

		if _, err := svc.Pay(context.Background(), req.PaymentRequestID, "payer"); !errors.Is(err, port.ErrPaymentRequestExpired) {
			t.Fatalf("err = %v, want ErrPaymentRequestExpired", err)
		}
	})

	t.Run("cancel prevents further payment", func(t *testing.T) {
		s, svc := newPaymentRequestTestDeps()
		s.wallets["requester"] = &model.Wallet{WalletID: "requester", UserID: "u1", Currency: "USD", Status: model.WalletActive}
		s.wallets["payer"] = &model.Wallet{WalletID: "payer", UserID: "u2", Currency: "USD", Status: model.WalletActive, Balance: 1000}

		req, _ := svc.CreatePaymentRequest(context.Background(), "requester", nil, 100, "", 0)
		if err := svc.Cancel(context.Background(), req.PaymentRequestID); err != nil {
			t.Fatalf("cancel: %v", err)
		}
		if _, err := svc.Pay(context.Background(), req.PaymentRequestID, "payer"); !errors.Is(err, port.ErrPaymentRequestNotPending) {
			t.Fatalf("err = %v, want ErrPaymentRequestNotPending", err)
		}
	})
}
