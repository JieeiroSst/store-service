package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

func TestReverse(t *testing.T) {
	t.Run("reverses a deposit back out of the wallet", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 0)

		deposit, err := d.svc.Deposit(context.Background(), "w1", 500, "", "")
		if err != nil {
			t.Fatalf("deposit: %v", err)
		}

		reversal, err := d.svc.Reverse(context.Background(), deposit.TransactionID, "duplicate top-up")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reversal.Type != model.TxnReversal {
			t.Errorf("type = %v, want reversal", reversal.Type)
		}
		if got := d.state.wallets["w1"].Balance; got != 0 {
			t.Errorf("balance = %d, want 0", got)
		}
		if d.state.transactions[deposit.TransactionID].Status != model.TxnReversed {
			t.Error("original deposit should be marked REVERSED")
		}
	})

	t.Run("reverses a withdrawal back into the wallet", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 1000)

		withdraw, err := d.svc.Withdraw(context.Background(), "w1", 300, "", "")
		if err != nil {
			t.Fatalf("withdraw: %v", err)
		}
		if _, err := d.svc.Reverse(context.Background(), withdraw.TransactionID, "wrong account"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := d.state.wallets["w1"].Balance; got != 1000 {
			t.Errorf("balance = %d, want 1000", got)
		}
	})

	t.Run("rejects reversing the same transaction twice", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("w1", "u1", "USD", 500)

		deposit, _ := d.svc.Deposit(context.Background(), "w1", 500, "", "")
		if _, err := d.svc.Reverse(context.Background(), deposit.TransactionID, "r1"); err != nil {
			t.Fatalf("first reverse: %v", err)
		}
		if _, err := d.svc.Reverse(context.Background(), deposit.TransactionID, "r2"); !errors.Is(err, port.ErrTransactionNotReversible) {
			t.Fatalf("err = %v, want ErrTransactionNotReversible", err)
		}
	})

	t.Run("rejects reversing a transfer leg directly", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 1000)
		d.seedWallet("receiver", "u2", "USD", 0)

		transfer, err := d.svc.Transfer(context.Background(), "sender", "receiver", 200, "", "")
		if err != nil {
			t.Fatalf("transfer: %v", err)
		}

		if _, err := d.svc.Reverse(context.Background(), transfer.OutTransactionID, "oops"); !errors.Is(err, port.ErrTransactionNotReversible) {
			t.Fatalf("err = %v, want ErrTransactionNotReversible", err)
		}
	})
}

func TestReverseTransfer(t *testing.T) {
	t.Run("returns the money to the sender", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 1000)
		d.seedWallet("receiver", "u2", "USD", 0)

		transfer, err := d.svc.Transfer(context.Background(), "sender", "receiver", 300, "", "")
		if err != nil {
			t.Fatalf("transfer: %v", err)
		}

		reversed, err := d.svc.ReverseTransfer(context.Background(), transfer.TransferID, "accidental transfer")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reversed.Status != model.TxnReversed {
			t.Errorf("status = %v, want REVERSED", reversed.Status)
		}
		if got := d.state.wallets["sender"].Balance; got != 1000 {
			t.Errorf("sender balance = %d, want 1000", got)
		}
		if got := d.state.wallets["receiver"].Balance; got != 0 {
			t.Errorf("receiver balance = %d, want 0", got)
		}
	})

	t.Run("fails if the receiver already spent the money", func(t *testing.T) {
		d := newTransactionTestDeps()
		d.seedWallet("sender", "u1", "USD", 1000)
		d.seedWallet("receiver", "u2", "USD", 0)

		transfer, _ := d.svc.Transfer(context.Background(), "sender", "receiver", 300, "", "")
		if _, err := d.svc.Withdraw(context.Background(), "receiver", 300, "", ""); err != nil {
			t.Fatalf("receiver spends the money: %v", err)
		}

		if _, err := d.svc.ReverseTransfer(context.Background(), transfer.TransferID, "too late"); !errors.Is(err, port.ErrInsufficientBalance) {
			t.Fatalf("err = %v, want ErrInsufficientBalance", err)
		}
	})
}
