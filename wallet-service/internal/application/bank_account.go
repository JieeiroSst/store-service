package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type bankAccountService struct {
	core
}

func NewBankAccountService(d Deps) port.BankAccountService {
	return &bankAccountService{core: newCore(d)}
}

func (s *bankAccountService) Link(ctx context.Context, cmd port.LinkBankAccountCommand) (*model.BankAccount, error) {
	if cmd.UserID == uuid.Nil {
		return nil, model.Invalid("user_id is required")
	}
	bank, err := s.Banks.GetByID(ctx, cmd.BankID)
	if err != nil {
		return nil, err
	}
	account, err := model.NewBankAccount(cmd.UserID, bank.ID, cmd.AccountNumber, cmd.AccountName, s.now())
	if err != nil {
		return nil, err
	}
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		existing, err := s.BankAccounts.ListActiveByUser(ctx, cmd.UserID)
		if err != nil {
			return err
		}
		account.IsDefault = len(existing) == 0
		return s.BankAccounts.Create(ctx, account)
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *bankAccountService) List(ctx context.Context, userID uuid.UUID) ([]*model.BankAccount, error) {
	return s.BankAccounts.ListActiveByUser(ctx, userID)
}

func (s *bankAccountService) Remove(ctx context.Context, userID, accountID uuid.UUID) error {
	return s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		account, err := s.owned(ctx, userID, accountID)
		if err != nil {
			return err
		}
		wasDefault := account.IsDefault
		account.Remove(s.now())
		if err := s.BankAccounts.Update(ctx, account); err != nil {
			return err
		}
		if !wasDefault {
			return nil
		}
		remaining, err := s.BankAccounts.ListActiveByUser(ctx, userID)
		if err != nil || len(remaining) == 0 {
			return err
		}
		next := remaining[0]
		next.IsDefault = true
		next.UpdatedAt = s.now()
		return s.BankAccounts.Update(ctx, next)
	})
}

func (s *bankAccountService) SetDefault(ctx context.Context, userID, accountID uuid.UUID) (*model.BankAccount, error) {
	var out *model.BankAccount
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		account, err := s.owned(ctx, userID, accountID)
		if err != nil {
			return err
		}
		if err := s.BankAccounts.ClearDefault(ctx, userID); err != nil {
			return err
		}
		account.IsDefault = true
		account.UpdatedAt = s.now()
		out = account
		return s.BankAccounts.Update(ctx, account)
	})
	return out, err
}

func (s *bankAccountService) owned(ctx context.Context, userID, accountID uuid.UUID) (*model.BankAccount, error) {
	account, err := s.BankAccounts.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := account.EnsureUsableBy(userID); err != nil {
		return nil, err
	}
	return account, nil
}
