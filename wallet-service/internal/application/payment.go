package application

import (
	"context"

	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type paymentService struct {
	core
}

func NewPaymentService(d Deps) port.PaymentService {
	return &paymentService{core: newCore(d)}
}

func (s *paymentService) Pay(ctx context.Context, cmd port.PayCommand) (*model.Transaction, error) {
	if err := validateIdempotencyKey(cmd.IdempotencyKey); err != nil {
		return nil, err
	}
	w, err := s.ownedWallet(ctx, cmd.WalletID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypePayment); err != nil || prior != nil {
		return prior, err
	}
	if err := model.ValidateAmount(cmd.Amount, w.Currency); err != nil {
		return nil, err
	}
	merchant, err := s.Merchants.GetByID(ctx, cmd.MerchantID)
	if err != nil {
		return nil, err
	}
	if err := s.verifyPin(ctx, cmd.WalletID, cmd.UserID, cmd.Pin); err != nil {
		return nil, err
	}

	var tx *model.Transaction
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		locked, err := s.Wallets.GetByIDForUpdate(ctx, cmd.WalletID)
		if err != nil {
			return err
		}
		if err := locked.EnsureActive(); err != nil {
			return err
		}
		if err := s.checkLimits(ctx, locked, cmd.Amount); err != nil {
			return err
		}
		now := s.now()
		if err := locked.Debit(cmd.Amount, now); err != nil {
			return err
		}
		tx = model.NewTransaction(locked, model.TransactionTypePayment, model.TransactionStatusCompleted, cmd.Amount, now)
		tx.IdempotencyKey = cmd.IdempotencyKey
		tx.MerchantID = model.NullID(merchant.ID)
		tx.AcquirerBankID = model.NullID(merchant.AcquirerBankID)
		tx.Description = cmd.Description
		tx.Metadata = cmd.Metadata
		return s.persistDebit(ctx, locked, tx, cmd.Amount, now)
	})
	if err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypePayment)
	}
	s.invalidateWallets(ctx, cmd.WalletID)
	s.Logger.Info("wallet payment completed", zap.String("tx_id", tx.ID.String()), zap.String("merchant_id", merchant.ID.String()))
	return tx, nil
}

func (s *paymentService) Refund(ctx context.Context, cmd port.RefundCommand) (*model.Transaction, error) {
	if err := validateIdempotencyKey(cmd.IdempotencyKey); err != nil {
		return nil, err
	}
	original, err := s.Transactions.GetByID(ctx, cmd.TransactionID)
	if err != nil {
		return nil, err
	}
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, original.WalletID, model.TransactionTypeRefund); err != nil || prior != nil {
		return prior, err
	}
	if !cmd.Amount.IsZero() {
		if err := model.ValidateAmount(cmd.Amount, original.Currency); err != nil {
			return nil, err
		}
	}

	var refund *model.Transaction
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		orig, err := s.Transactions.GetByIDForUpdate(ctx, cmd.TransactionID)
		if err != nil {
			return err
		}
		w, err := s.Wallets.GetByIDForUpdate(ctx, orig.WalletID)
		if err != nil {
			return err
		}
		now := s.now()
		amount, err := orig.ApplyRefund(cmd.Amount, now)
		if err != nil {
			return err
		}
		if err := w.Credit(amount, now); err != nil {
			return err
		}
		refund = model.NewTransaction(w, model.TransactionTypeRefund, model.TransactionStatusCompleted, amount, now)
		refund.IdempotencyKey = cmd.IdempotencyKey
		refund.ReferenceID = model.NullID(orig.ID)
		refund.MerchantID = orig.MerchantID
		refund.AcquirerBankID = orig.AcquirerBankID
		refund.Description = cmd.Reason
		if err := s.persistCredit(ctx, w, refund, amount, now); err != nil {
			return err
		}
		return s.Transactions.Update(ctx, orig)
	})
	if err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, original.WalletID, model.TransactionTypeRefund)
	}
	s.invalidateWallets(ctx, original.WalletID)
	s.invalidateTransactions(ctx, original.ID)
	s.Logger.Info("payment refunded", zap.String("refund_id", refund.ID.String()), zap.String("payment_id", original.ID.String()), zap.String("amount", refund.Amount.String()))
	return refund, nil
}
