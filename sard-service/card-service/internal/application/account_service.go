package application

import (
	"context"
	"crypto/rand"
	"io"
	"strings"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/JIeeiroSst/card-service/internal/port"
)

type AccountService struct {
	catalog      *domain.Catalog
	accounts     port.AccountRepository
	cards        port.CardRepository
	transactions port.TransactionRepository
	customers    port.CustomerDirectory
	locker       port.Locker
	clock        port.Clock
	metrics      port.Metrics
	random       io.Reader
}

func NewAccountService(catalog *domain.Catalog, accounts port.AccountRepository, cards port.CardRepository, transactions port.TransactionRepository, customers port.CustomerDirectory, locker port.Locker, clock port.Clock, metrics port.Metrics) *AccountService {
	return &AccountService{
		catalog: catalog, accounts: accounts, cards: cards, transactions: transactions,
		customers: customers, locker: locker, clock: clock, metrics: metrics, random: rand.Reader,
	}
}

type OpenAccountCommand struct {
	CustomerID  string
	ProgramCode string
	CreditLimit int64
}

func (s *AccountService) Programs() []domain.Program { return s.catalog.All() }

func (s *AccountService) Open(ctx context.Context, cmd OpenAccountCommand) (*domain.Account, error) {
	if err := domain.ValidateID("customer_id", cmd.CustomerID); err != nil {
		return nil, err
	}
	program, err := s.catalog.Get(cmd.ProgramCode)
	if err != nil {
		return nil, err
	}
	customer, err := s.customers.GetCustomer(ctx, cmd.CustomerID)
	if err != nil {
		return nil, err
	}
	var out *domain.Account
	err = s.locker.WithLock(ctx, customerLockKey(cmd.CustomerID), func(ctx context.Context) error {
		existing, err := s.accounts.ListByCustomer(ctx, cmd.CustomerID)
		if err != nil {
			return err
		}
		for _, a := range existing {
			if a.ProgramCode == program.Code && a.Status != domain.AccountCancelled {
				return domain.Conflict("customer already has account %s in program %s", a.ID, program.Code)
			}
		}
		id, err := newID(s.random)
		if err != nil {
			return err
		}
		a, err := domain.OpenAccount(id, customer, program, cmd.CreditLimit, s.clock.Now())
		if err != nil {
			return err
		}
		out = a
		return s.accounts.Create(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveLifecycle("account_open")
	return out, nil
}

func (s *AccountService) Get(ctx context.Context, id string) (*domain.Account, error) {
	return s.accounts.Get(ctx, id)
}

func (s *AccountService) ListByCustomer(ctx context.Context, customerID string) ([]*domain.Account, error) {
	if err := domain.ValidateID("customer_id", customerID); err != nil {
		return nil, err
	}
	return s.accounts.ListByCustomer(ctx, customerID)
}

func (s *AccountService) Block(ctx context.Context, id, reason string) (*domain.Account, error) {
	return s.mutate(ctx, id, "account_block", func(a *domain.Account, now time.Time) error { return a.Block(reason, now) })
}

func (s *AccountService) Unblock(ctx context.Context, id string) (*domain.Account, error) {
	return s.mutate(ctx, id, "account_unblock", func(a *domain.Account, now time.Time) error { return a.Unblock(now) })
}

func (s *AccountService) SetCreditLimit(ctx context.Context, id string, limit int64) (*domain.Account, error) {
	return s.mutate(ctx, id, "credit_limit", func(a *domain.Account, now time.Time) error {
		p, err := s.catalog.Get(a.ProgramCode)
		if err != nil {
			return err
		}
		return a.SetCreditLimit(limit, p, now)
	})
}

func (s *AccountService) Cancel(ctx context.Context, id, reason string) (*domain.Account, error) {
	return s.mutate(ctx, id, "account_cancel", func(a *domain.Account, now time.Time) error {
		if err := a.Cancel(reason, now); err != nil {
			return err
		}
		cards, err := s.cards.ListByAccount(ctx, a.ID)
		if err != nil {
			return err
		}
		for _, c := range cards {
			if c.Status.Terminal() {
				continue
			}
			err := s.locker.WithLock(ctx, cardLockKey(c.ID), func(ctx context.Context) error {
				fresh, err := s.cards.Get(ctx, c.ID)
				if err != nil {
					return err
				}
				if fresh.Status.Terminal() {
					return nil
				}
				if err := fresh.Cancel("account cancelled", now); err != nil {
					return err
				}
				return s.cards.Update(ctx, fresh)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AccountService) ReceivePayment(ctx context.Context, id string, amount int64, description string) (*domain.Account, *domain.Transaction, error) {
	description = strings.TrimSpace(description)
	if len(description) > 140 {
		return nil, nil, domain.Invalid("description must be at most 140 characters")
	}
	var txn *domain.Transaction
	a, err := s.mutate(ctx, id, "payment", func(a *domain.Account, now time.Time) error {
		if err := a.ReceivePayment(amount, now); err != nil {
			return err
		}
		tid, err := newID(s.random)
		if err != nil {
			return err
		}
		txn = &domain.Transaction{
			ID: tid, AccountID: a.ID, Type: domain.TxnPayment, ProcessingCode: domain.ProcessingPayment,
			Amount: amount, BalanceAfter: a.Balance, Description: description, CreatedAt: now,
		}
		return s.transactions.Create(ctx, txn)
	})
	if err != nil {
		return nil, nil, err
	}
	return a, txn, nil
}

func (s *AccountService) Transactions(ctx context.Context, id string, limit int) ([]*domain.Transaction, error) {
	if _, err := s.accounts.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.transactions.ListByAccount(ctx, id, clampLimit(limit))
}

func (s *AccountService) mutate(ctx context.Context, id, event string, fn func(a *domain.Account, now time.Time) error) (*domain.Account, error) {
	var out *domain.Account
	err := s.locker.WithLock(ctx, accountLockKey(id), func(ctx context.Context) error {
		a, err := s.accounts.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(a, s.clock.Now()); err != nil {
			return err
		}
		out = a
		return s.accounts.Update(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveLifecycle(event)
	return out, nil
}
