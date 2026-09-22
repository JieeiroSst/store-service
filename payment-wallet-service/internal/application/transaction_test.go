package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

type transactionTestDeps struct {
	state *fakeState
	svc   port.TransactionUsecase
}

func newTransactionTestDeps() *transactionTestDeps {
	s := newFakeState()
	return &transactionTestDeps{
		state: s,
		svc:   NewTransactionService(&fakeWalletRepo{s: s}, &fakeTransactionRepo{s: s}, &fakeTransferRepo{s: s}),
	}
}

func (d *transactionTestDeps) seedWallet(id, userID, currency string, balance int64) {
	d.state.wallets[id] = &model.Wallet{WalletID: id, UserID: userID, Currency: currency, Balance: balance, Status: model.WalletActive}
}

func TestDeposit(t *testing.T) {
	t.Run("increases the balance and records a completed transaction", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 1000)

		txn, err := d.svc.Deposit(context.Background(), "w1", 500, "ref-1", "top up")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if txn.Type != model.TxnDeposit || txn.Status != model.TxnCompleted {
			t.Errorf("txn = %+v, want completed deposit", txn)
		}
		if got := d.state.wallets["w1"].Balance; got != 1500 {
			t.Errorf("balance = %d, want 1500", got)
		}
	})

	t.Run("replays the same transaction instead of depositing twice", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 0)

		first, err := d.svc.Deposit(context.Background(), "w1", 500, "ref-1", "")
		if err != nil {
			t.Fatalf("first deposit: %v", err)
		}
		second, err := d.svc.Deposit(context.Background(), "w1", 500, "ref-1", "")
		if err != nil {
			t.Fatalf("second deposit: %v", err)
		}
		if first.TransactionID != second.TransactionID {
			t.Errorf("expected idempotent replay, got two different transactions")
		}
		if got := d.state.wallets["w1"].Balance; got != 500 {
			t.Errorf("balance = %d, want 500 (only credited once)", got)
		}
	})

	t.Run("rejects a non-positive amount", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 0)

		if _, err := d.svc.Deposit(context.Background(), "w1", 0, "", ""); !errors.Is(err, port.ErrInvalidAmount) {
			t.Fatalf("err = %v, want ErrInvalidAmount", err)
		}
	})
}

func TestWithdraw(t *testing.T) {
	t.Run("decreases the balance", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 1000)

		txn, err := d.svc.Withdraw(context.Background(), "w1", 400, "ref-w1", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if txn.Type != model.TxnWithdraw {
			t.Errorf("type = %v, want withdraw", txn.Type)
		}
		if got := d.state.wallets["w1"].Balance; got != 600 {
			t.Errorf("balance = %d, want 600", got)
		}
	})

	t.Run("rejects withdrawing more than the balance", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 100)

		if _, err := d.svc.Withdraw(context.Background(), "w1", 200, "", ""); !errors.Is(err, port.ErrInsufficientBalance) {
			t.Fatalf("err = %v, want ErrInsufficientBalance", err)
		}
		if got := d.state.wallets["w1"].Balance; got != 100 {
			t.Errorf("balance = %d, want unchanged 100", got)
		}
	})

	t.Run("rejects withdrawing from a frozen wallet", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 100)
		d.state.wallets["w1"].Status = model.WalletFrozen

		if _, err := d.svc.Withdraw(context.Background(), "w1", 50, "", ""); !errors.Is(err, port.ErrWalletNotActive) {
			t.Fatalf("err = %v, want ErrWalletNotActive", err)
		}
	})
}

func TestTransfer(t *testing.T) {
	t.Run("moves the balance between two wallets", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 1000)
		d.seedWallet("receiver", "u2", "USD", 0)

		transfer, err := d.svc.Transfer(context.Background(), "sender", "receiver", 300, "tr-1", "lunch")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if transfer.Amount != 300 {
			t.Errorf("amount = %d, want 300", transfer.Amount)
		}
		if got := d.state.wallets["sender"].Balance; got != 700 {
			t.Errorf("sender balance = %d, want 700", got)
		}
		if got := d.state.wallets["receiver"].Balance; got != 300 {
			t.Errorf("receiver balance = %d, want 300", got)
		}
	})

	t.Run("rejects transferring to the same wallet", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 1000)

		if _, err := d.svc.Transfer(context.Background(), "w1", "w1", 100, "", ""); !errors.Is(err, port.ErrSameWallet) {
			t.Fatalf("err = %v, want ErrSameWallet", err)
		}
	})

	t.Run("rejects a currency mismatch", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 1000)
		d.seedWallet("receiver", "u2", "EUR", 0)

		if _, err := d.svc.Transfer(context.Background(), "sender", "receiver", 100, "", ""); !errors.Is(err, port.ErrCurrencyMismatch) {
			t.Fatalf("err = %v, want ErrCurrencyMismatch", err)
		}
	})

	t.Run("rejects insufficient balance without touching either wallet", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 50)
		d.seedWallet("receiver", "u2", "USD", 0)

		if _, err := d.svc.Transfer(context.Background(), "sender", "receiver", 100, "", ""); !errors.Is(err, port.ErrInsufficientBalance) {
			t.Fatalf("err = %v, want ErrInsufficientBalance", err)
		}
		if got := d.state.wallets["sender"].Balance; got != 50 {
			t.Errorf("sender balance = %d, want unchanged 50", got)
		}
		if got := d.state.wallets["receiver"].Balance; got != 0 {
			t.Errorf("receiver balance = %d, want unchanged 0", got)
		}
	})
}
