package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

func TestCreateWallet(t *testing.T) {
	t.Run("creates a wallet with a zero balance", func(t *testing.T) {
		s := newFakeState()
		svc := NewWalletService(&fakeWalletRepo{s: s})

		wallet, err := svc.CreateWallet(context.Background(), "u1", "USD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wallet.Balance != 0 {
			t.Errorf("balance = %d, want 0", wallet.Balance)
		}
		if wallet.Currency != "USD" {
			t.Errorf("currency = %s, want USD", wallet.Currency)
		}
	})

	t.Run("rejects a second wallet for the same user", func(t *testing.T) {
		s := newFakeState()
		svc := NewWalletService(&fakeWalletRepo{s: s})

		if _, err := svc.CreateWallet(context.Background(), "u1", "USD"); err != nil {
			t.Fatalf("first create: %v", err)
		}
		if _, err := svc.CreateWallet(context.Background(), "u1", "USD"); !errors.Is(err, port.ErrWalletAlreadyExists) {
			t.Fatalf("err = %v, want ErrWalletAlreadyExists", err)
		}
	})
}

func TestWalletLifecycle(t *testing.T) {
	t.Run("freeze then unfreeze", func(t *testing.T) {
		s := newFakeState()
		s.wallets["w1"] = &model.Wallet{WalletID: "w1", UserID: "u1", Currency: "USD", Status: model.WalletActive}
		svc := NewWalletService(&fakeWalletRepo{s: s})

		frozen, err := svc.FreezeWallet(context.Background(), "w1", "suspected fraud")
		if err != nil {
			t.Fatalf("freeze: %v", err)
		}
		if frozen.Status != model.WalletFrozen {
			t.Errorf("status = %v, want FROZEN", frozen.Status)
		}

		unfrozen, err := svc.UnfreezeWallet(context.Background(), "w1")
		if err != nil {
			t.Fatalf("unfreeze: %v", err)
		}
		if unfrozen.Status != model.WalletActive {
			t.Errorf("status = %v, want ACTIVE", unfrozen.Status)
		}
	})

	t.Run("rejects freezing an already frozen wallet", func(t *testing.T) {
		s := newFakeState()
		s.wallets["w1"] = &model.Wallet{WalletID: "w1", UserID: "u1", Currency: "USD", Status: model.WalletFrozen}
		svc := NewWalletService(&fakeWalletRepo{s: s})

		if _, err := svc.FreezeWallet(context.Background(), "w1", "again"); !errors.Is(err, port.ErrInvalidWalletStatusTransition) {
			t.Fatalf("err = %v, want ErrInvalidWalletStatusTransition", err)
		}
	})

	t.Run("closes an empty wallet but rejects a non-empty one", func(t *testing.T) {
		s := newFakeState()
		s.wallets["empty"] = &model.Wallet{WalletID: "empty", UserID: "u1", Currency: "USD", Status: model.WalletActive, Balance: 0}
		s.wallets["funded"] = &model.Wallet{WalletID: "funded", UserID: "u2", Currency: "USD", Status: model.WalletActive, Balance: 100}
		svc := NewWalletService(&fakeWalletRepo{s: s})

		closed, err := svc.CloseWallet(context.Background(), "empty")
		if err != nil {
			t.Fatalf("close empty wallet: %v", err)
		}
		if closed.Status != model.WalletClosed {
			t.Errorf("status = %v, want CLOSED", closed.Status)
		}

		if _, err := svc.CloseWallet(context.Background(), "funded"); !errors.Is(err, port.ErrWalletNotEmpty) {
			t.Fatalf("err = %v, want ErrWalletNotEmpty", err)
		}
	})
}

func TestTransactionLimits(t *testing.T) {
	t.Run("rejects a withdrawal above the per-transaction limit", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 10000)
		d.state.wallets["w1"].PerTransactionLimit = 500

		if _, err := d.svc.Withdraw(context.Background(), "w1", 600, "", ""); !errors.Is(err, port.ErrPerTransactionLimitExceeded) {
			t.Fatalf("err = %v, want ErrPerTransactionLimitExceeded", err)
		}
	})

	t.Run("rejects a transfer that would exceed the sender's daily limit", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 10000)
		d.seedWallet("receiver", "u2", "USD", 0)
		d.state.wallets["sender"].DailyLimit = 1000

		if _, err := d.svc.Transfer(context.Background(), "sender", "receiver", 500, "", ""); err != nil {
			t.Fatalf("first transfer under the limit: %v", err)
		}
		if _, err := d.svc.Transfer(context.Background(), "sender", "receiver", 600, "", ""); !errors.Is(err, port.ErrDailyLimitExceeded) {
			t.Fatalf("err = %v, want ErrDailyLimitExceeded", err)
		}
	})
}
