package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type walletService struct {
	core
}

func NewWalletService(d Deps) port.WalletService {
	return &walletService{core: newCore(d)}
}

func (s *walletService) Create(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error) {
	w, err := model.NewWallet(userID, model.Currency(strings.ToUpper(string(currency))), s.now())
	if err != nil {
		return nil, err
	}
	if err := s.Wallets.Create(ctx, w); err != nil {
		return nil, err
	}
	s.Logger.Info("wallet created", zap.String("wallet_id", w.ID.String()), zap.String("currency", string(w.Currency)))
	return w, nil
}

func (s *walletService) Get(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error) {
	key := fmt.Sprintf(walletCacheKey, walletID)
	if w, ok := cacheGet[model.Wallet](ctx, &s.core, key); ok {
		return w, nil
	}
	w, err := s.Wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	cacheSet(ctx, &s.core, key, w, walletCacheTTL)
	return w, nil
}

func (s *walletService) GetByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error) {
	return s.Wallets.GetByUserAndCurrency(ctx, userID, model.Currency(strings.ToUpper(string(currency))))
}

func (s *walletService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Wallet, error) {
	return s.Wallets.ListByUser(ctx, userID)
}

func (s *walletService) Balance(ctx context.Context, walletID uuid.UUID) (*model.Balance, error) {
	w, err := s.Wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	b := w.Snapshot()
	return &b, nil
}

func (s *walletService) Credit(ctx context.Context, cmd port.CreditCommand) (*model.Transaction, error) {
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypeAdjustment); err != nil || prior != nil {
		return prior, err
	}
	var tx *model.Transaction
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		w, err := s.Wallets.GetByIDForUpdate(ctx, cmd.WalletID)
		if err != nil {
			return err
		}
		if err := w.EnsureActive(); err != nil {
			return err
		}
		if err := model.ValidateAmount(cmd.Amount, w.Currency); err != nil {
			return err
		}
		now := s.now()
		if err := w.Credit(cmd.Amount, now); err != nil {
			return err
		}
		tx = model.NewTransaction(w, model.TransactionTypeAdjustment, model.TransactionStatusCompleted, cmd.Amount, now)
		tx.IdempotencyKey = cmd.IdempotencyKey
		tx.Description = cmd.Description
		return s.persistCredit(ctx, w, tx, cmd.Amount, now)
	})
	if err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypeAdjustment)
	}
	s.invalidateWallets(ctx, cmd.WalletID)
	return tx, nil
}

func (s *walletService) mutate(ctx context.Context, walletID uuid.UUID, fn func(w *model.Wallet) error) (*model.Wallet, error) {
	var out *model.Wallet
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		w, err := s.Wallets.GetByIDForUpdate(ctx, walletID)
		if err != nil {
			return err
		}
		if err := fn(w); err != nil {
			return err
		}
		out = w
		return s.Wallets.Update(ctx, w)
	})
	if err != nil {
		return nil, err
	}
	s.invalidateWallets(ctx, walletID)
	return out, nil
}

func (s *walletService) Freeze(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error) {
	return s.mutate(ctx, walletID, func(w *model.Wallet) error { return w.Freeze(s.now()) })
}

func (s *walletService) Unfreeze(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error) {
	return s.mutate(ctx, walletID, func(w *model.Wallet) error { return w.Unfreeze(s.now()) })
}

func (s *walletService) Close(ctx context.Context, walletID, userID uuid.UUID) (*model.Wallet, error) {
	return s.mutate(ctx, walletID, func(w *model.Wallet) error {
		if userID != uuid.Nil {
			if err := w.EnsureOwner(userID); err != nil {
				return err
			}
		}
		return w.Close(s.now())
	})
}

func (s *walletService) ChangeTier(ctx context.Context, walletID uuid.UUID, tier model.Tier) (*model.Wallet, error) {
	tier = model.Tier(strings.ToUpper(string(tier)))
	return s.mutate(ctx, walletID, func(w *model.Wallet) error { return w.ChangeTier(tier, s.now()) })
}

func (s *walletService) Limits(ctx context.Context, walletID uuid.UUID) (*model.LimitUsage, error) {
	w, err := s.Wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	used, err := s.Ledger.SumDebitsSince(ctx, w.ID, model.StartOfDay(now))
	if err != nil {
		return nil, err
	}
	usage := model.NewLimitUsage(w, used, now)
	return &usage, nil
}

func (s *walletService) SetPin(ctx context.Context, cmd port.SetPinCommand) error {
	if err := model.ValidatePinFormat(cmd.NewPin); err != nil {
		return err
	}
	w, err := s.ownedWallet(ctx, cmd.WalletID, cmd.UserID)
	if err != nil {
		return err
	}
	if w.HasPin {
		if err := s.verifyPin(ctx, cmd.WalletID, cmd.UserID, cmd.CurrentPin); err != nil {
			return err
		}
	}
	hash, err := s.Pins.Hash(cmd.NewPin)
	if err != nil {
		return err
	}
	_, err = s.mutate(ctx, cmd.WalletID, func(w *model.Wallet) error {
		w.SetPinHash(hash, s.now())
		return nil
	})
	if err == nil {
		s.Logger.Info("wallet pin updated", zap.String("wallet_id", cmd.WalletID.String()))
	}
	return err
}
