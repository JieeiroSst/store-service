package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

func newPocketTestDeps() (*fakeState, port.PocketUsecase) {
	s := newFakeState()
	return s, NewPocketService(&fakeWalletRepo{s: s}, &fakePocketRepo{s: s})
}

func TestPocket(t *testing.T) {
	t.Run("deposit moves money from wallet into pocket", func(t *testing.T) {
		s, svc := newPocketTestDeps()
		s.wallets["w1"] = &model.Wallet{WalletID: "w1", UserID: "u1", Currency: "USD", Status: model.WalletActive, Balance: 1000}

		pocket, err := svc.CreatePocket(context.Background(), "w1", "savings")
		if err != nil {
			t.Fatalf("create pocket: %v", err)
		}

		updated, err := svc.DepositToPocket(context.Background(), pocket.PocketID, 300)
		if err != nil {
			t.Fatalf("deposit to pocket: %v", err)
		}
		if updated.Balance != 300 {
			t.Errorf("pocket balance = %d, want 300", updated.Balance)
		}
		if got := s.wallets["w1"].Balance; got != 700 {
			t.Errorf("wallet balance = %d, want 700", got)
		}
	})

	t.Run("withdraw moves money back from pocket into wallet", func(t *testing.T) {
		s, svc := newPocketTestDeps()
		s.wallets["w1"] = &model.Wallet{WalletID: "w1", UserID: "u1", Currency: "USD", Status: model.WalletActive, Balance: 1000}
		pocket, _ := svc.CreatePocket(context.Background(), "w1", "savings")
		if _, err := svc.DepositToPocket(context.Background(), pocket.PocketID, 300); err != nil {
			t.Fatalf("deposit to pocket: %v", err)
		}

		updated, err := svc.WithdrawFromPocket(context.Background(), pocket.PocketID, 100)
		if err != nil {
			t.Fatalf("withdraw from pocket: %v", err)
		}
		if updated.Balance != 200 {
			t.Errorf("pocket balance = %d, want 200", updated.Balance)
		}
		if got := s.wallets["w1"].Balance; got != 800 {
			t.Errorf("wallet balance = %d, want 800", got)
		}
	})

	t.Run("rejects withdrawing more than the pocket holds", func(t *testing.T) {
		s, svc := newPocketTestDeps()
		s.wallets["w1"] = &model.Wallet{WalletID: "w1", UserID: "u1", Currency: "USD", Status: model.WalletActive, Balance: 1000}
		pocket, _ := svc.CreatePocket(context.Background(), "w1", "savings")

		if _, err := svc.WithdrawFromPocket(context.Background(), pocket.PocketID, 50); !errors.Is(err, port.ErrInsufficientBalance) {
			t.Fatalf("err = %v, want ErrInsufficientBalance", err)
		}
	})

	t.Run("closing a pocket returns its remaining balance to the wallet", func(t *testing.T) {
		s, svc := newPocketTestDeps()
		s.wallets["w1"] = &model.Wallet{WalletID: "w1", UserID: "u1", Currency: "USD", Status: model.WalletActive, Balance: 1000}
		pocket, _ := svc.CreatePocket(context.Background(), "w1", "savings")
		if _, err := svc.DepositToPocket(context.Background(), pocket.PocketID, 300); err != nil {
			t.Fatalf("deposit to pocket: %v", err)
		}

		wallet, err := svc.ClosePocket(context.Background(), pocket.PocketID)
		if err != nil {
			t.Fatalf("close pocket: %v", err)
		}
		if wallet.Balance != 1000 {
			t.Errorf("wallet balance = %d, want 1000 (pocket balance returned)", wallet.Balance)
		}
		if _, ok := s.pockets[pocket.PocketID]; ok {
			t.Error("pocket should be removed after closing")
		}
	})
}
