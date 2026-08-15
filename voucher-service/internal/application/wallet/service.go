package wallet

import (
	"context"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/wallet"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
)

type Service struct {
	repo      WalletRepository
	ledger    LedgerRepository
	txManager txmanager.TxManager
	clock     shared.Clock
}

func NewService(repo WalletRepository, ledger LedgerRepository, txManager txmanager.TxManager, clock shared.Clock) WalletService {
	return &Service{repo: repo, ledger: ledger, txManager: txManager, clock: clock}
}

func (s *Service) GetOrCreateWallet(ctx context.Context, ownerType wallet.OwnerType, ownerID, currency string) (*wallet.Wallet, error) {
	w, err := s.repo.FindByOwner(ctx, ownerType, ownerID)
	if err == nil {
		return w, nil
	}
	w, err = wallet.NewWallet(ownerType, ownerID, currency, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Service) GetBalance(ctx context.Context, ownerType wallet.OwnerType, ownerID string) (shared.Money, error) {
	w, err := s.repo.FindByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return shared.Money{}, err
	}
	return w.Balance, nil
}

func (s *Service) Credit(ctx context.Context, ownerType wallet.OwnerType, ownerID string, amount shared.Money, refType, refID, idempotencyKey string) error {
	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		w, err := s.repo.FindByOwnerForUpdate(ctx, ownerType, ownerID)
		if err != nil {
			return err
		}
		if err := w.Credit(amount, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, w); err != nil {
			return err
		}
		return s.ledger.Append(ctx, LedgerEntry{
			WalletID: w.ID, Type: "credit", Amount: amount, BalanceAfter: w.Balance,
			RefType: refType, RefID: refID, IdempotencyKey: idempotencyKey,
		})
	})
}

func (s *Service) Debit(ctx context.Context, ownerType wallet.OwnerType, ownerID string, amount shared.Money, refType, refID, idempotencyKey string) error {
	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		w, err := s.repo.FindByOwnerForUpdate(ctx, ownerType, ownerID)
		if err != nil {
			return err
		}
		if err := w.Debit(amount, s.clock.Now()); err != nil {
			return err
		}
		if err := s.repo.Save(ctx, w); err != nil {
			return err
		}
		return s.ledger.Append(ctx, LedgerEntry{
			WalletID: w.ID, Type: "debit", Amount: amount, BalanceAfter: w.Balance,
			RefType: refType, RefID: refID, IdempotencyKey: idempotencyKey,
		})
	})
}
