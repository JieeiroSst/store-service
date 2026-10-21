package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type fundingService struct {
	core
}

func NewFundingService(d Deps) port.FundingService {
	return &fundingService{core: newCore(d)}
}

func (s *fundingService) TopUp(ctx context.Context, cmd port.TopUpCommand) (*model.Transaction, error) {
	if err := validateIdempotencyKey(cmd.IdempotencyKey); err != nil {
		return nil, err
	}
	w, err := s.ownedWallet(ctx, cmd.WalletID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypeTopUp); err != nil || prior != nil {
		return prior, err
	}
	if err := w.EnsureActive(); err != nil {
		return nil, err
	}
	if err := model.ValidateAmount(cmd.Amount, w.Currency); err != nil {
		return nil, err
	}
	account, err := s.usableAccount(ctx, cmd.BankAccountID, cmd.UserID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	tx := model.NewTransaction(w, model.TransactionTypeTopUp, model.TransactionStatusPending, cmd.Amount, now)
	tx.IdempotencyKey = cmd.IdempotencyKey
	tx.BankAccountID = model.NullID(account.ID)
	tx.Description = cmd.Description
	if err := s.Transactions.Create(ctx, tx); err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypeTopUp)
	}

	ref, bankErr := s.BankGateway.Collect(ctx, port.BankTransfer{Reference: tx.ID, Account: account, Amount: cmd.Amount, Currency: w.Currency})
	if bankErr != nil {
		s.Logger.Warn("top-up collection failed", zap.String("tx_id", tx.ID.String()), zap.Error(bankErr))
		if err := s.failPending(ctx, tx, bankErr); err != nil {
			return nil, err
		}
		return nil, model.ErrBankDeclined
	}

	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		locked, err := s.Wallets.GetByIDForUpdate(ctx, w.ID)
		if err != nil {
			return err
		}
		current, err := s.Transactions.GetByIDForUpdate(ctx, tx.ID)
		if err != nil {
			return err
		}
		now := s.now()
		if err := current.Complete(ref, now); err != nil {
			return err
		}
		if err := locked.Credit(current.Amount, now); err != nil {
			return err
		}
		if err := s.Ledger.Append(ctx, model.NewLedgerEntry(locked, current.ID, model.EntryCredit, current.Amount, now)); err != nil {
			return err
		}
		if err := s.Wallets.Update(ctx, locked); err != nil {
			return err
		}
		tx = current
		return s.Transactions.Update(ctx, current)
	})
	if err != nil {
		s.Logger.Error("top-up collected but credit failed, needs reconciliation", zap.String("tx_id", tx.ID.String()), zap.String("bank_ref", ref), zap.Error(err))
		return nil, err
	}
	s.invalidateWallets(ctx, w.ID)
	s.Logger.Info("wallet topped up", zap.String("tx_id", tx.ID.String()), zap.String("amount", tx.Amount.String()))
	return tx, nil
}

func (s *fundingService) Withdraw(ctx context.Context, cmd port.WithdrawCommand) (*model.Transaction, error) {
	if err := validateIdempotencyKey(cmd.IdempotencyKey); err != nil {
		return nil, err
	}
	w, err := s.ownedWallet(ctx, cmd.WalletID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if prior, err := s.replay(ctx, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypeWithdrawal); err != nil || prior != nil {
		return prior, err
	}
	if err := model.ValidateAmount(cmd.Amount, w.Currency); err != nil {
		return nil, err
	}
	account, err := s.usableAccount(ctx, cmd.BankAccountID, cmd.UserID)
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
		now := s.now()
		fee := s.Fees.Withdrawal(cmd.Amount, locked.Currency)
		tx = model.NewTransaction(locked, model.TransactionTypeWithdrawal, model.TransactionStatusPending, cmd.Amount, now)
		tx.IdempotencyKey = cmd.IdempotencyKey
		tx.BankAccountID = model.NullID(account.ID)
		tx.Fee = fee
		tx.Description = cmd.Description
		total := tx.TotalDebit()
		if err := s.checkLimits(ctx, locked, total); err != nil {
			return err
		}
		if err := locked.Debit(total, now); err != nil {
			return err
		}
		return s.persistDebit(ctx, locked, tx, total, now)
	})
	if err != nil {
		return s.resolveDuplicate(ctx, err, cmd.IdempotencyKey, cmd.WalletID, model.TransactionTypeWithdrawal)
	}
	s.invalidateWallets(ctx, cmd.WalletID)

	ref, bankErr := s.BankGateway.Payout(ctx, port.BankTransfer{Reference: tx.ID, Account: account, Amount: tx.Amount, Currency: tx.Currency})
	if bankErr != nil {
		s.Logger.Warn("withdrawal payout failed, reversing", zap.String("tx_id", tx.ID.String()), zap.Error(bankErr))
		if err := s.reverseWithdrawal(ctx, tx, bankErr); err != nil {
			s.Logger.Error("withdrawal reversal failed, needs reconciliation", zap.String("tx_id", tx.ID.String()), zap.Error(err))
			return nil, err
		}
		return nil, model.ErrBankDeclined
	}

	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.Transactions.GetByIDForUpdate(ctx, tx.ID)
		if err != nil {
			return err
		}
		if err := current.Complete(ref, s.now()); err != nil {
			return err
		}
		tx = current
		return s.Transactions.Update(ctx, current)
	})
	if err != nil {
		s.Logger.Error("withdrawal paid out but not marked completed", zap.String("tx_id", tx.ID.String()), zap.String("bank_ref", ref), zap.Error(err))
		return nil, err
	}
	s.invalidateTransactions(ctx, tx.ID)
	s.Logger.Info("wallet withdrawal completed", zap.String("tx_id", tx.ID.String()), zap.String("amount", tx.Amount.String()))
	return tx, nil
}

func (s *fundingService) usableAccount(ctx context.Context, accountID, userID uuid.UUID) (*model.BankAccount, error) {
	account, err := s.BankAccounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := account.EnsureUsableBy(userID); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *fundingService) failPending(ctx context.Context, tx *model.Transaction, cause error) error {
	return s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		current, err := s.Transactions.GetByIDForUpdate(ctx, tx.ID)
		if err != nil {
			return err
		}
		if err := current.Fail(failureReason(cause), s.now()); err != nil {
			return err
		}
		return s.Transactions.Update(ctx, current)
	})
}

func (s *fundingService) reverseWithdrawal(ctx context.Context, tx *model.Transaction, cause error) error {
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		locked, err := s.Wallets.GetByIDForUpdate(ctx, tx.WalletID)
		if err != nil {
			return err
		}
		current, err := s.Transactions.GetByIDForUpdate(ctx, tx.ID)
		if err != nil {
			return err
		}
		now := s.now()
		if err := current.Fail(failureReason(cause), now); err != nil {
			return err
		}
		total := current.TotalDebit()
		if err := locked.Credit(total, now); err != nil {
			return err
		}
		if err := s.Ledger.Append(ctx, model.NewLedgerEntry(locked, current.ID, model.EntryCredit, total, now)); err != nil {
			return err
		}
		if err := s.Wallets.Update(ctx, locked); err != nil {
			return err
		}
		return s.Transactions.Update(ctx, current)
	})
	if err == nil {
		s.invalidateWallets(ctx, tx.WalletID)
	}
	return err
}

func failureReason(err error) string {
	var e *model.Error
	if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}
