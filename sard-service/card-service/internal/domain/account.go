package domain

import (
	"strings"
	"time"
)

type AccountStatus string

const (
	AccountNormal    AccountStatus = "NORMAL"
	AccountBlocked   AccountStatus = "BLOCKED"
	AccountCancelled AccountStatus = "CANCELLED"
)

type Customer struct {
	ID       string
	UserID   int64
	FullName string
	Eligible bool
	Reasons  []string
}

type Account struct {
	ID           string
	CustomerID   string
	UserID       int64
	ProgramCode  string
	Mode         Mode
	Currency     string
	HolderName   string
	Status       AccountStatus
	StatusReason string
	CreditLimit  int64
	Balance      int64
	Held         int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func OpenAccount(id string, c Customer, p Program, creditLimit int64, now time.Time) (*Account, error) {
	if !c.Eligible {
		return nil, Conflict("customer is not eligible for a card account: %s", strings.Join(c.Reasons, "; "))
	}
	a := &Account{
		ID:          id,
		CustomerID:  c.ID,
		UserID:      c.UserID,
		ProgramCode: p.Code,
		Mode:        p.Mode,
		Currency:    p.Currency,
		HolderName:  c.FullName,
		Status:      AccountNormal,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if p.Mode == ModeCredit {
		if creditLimit == 0 {
			creditLimit = p.DefaultCreditLimit
		}
		if err := a.SetCreditLimit(creditLimit, p, now); err != nil {
			return nil, err
		}
	} else if creditLimit != 0 {
		return nil, Invalid("credit_limit only applies to CREDIT programs")
	}
	return a, nil
}

func (a *Account) Available() int64 {
	return a.Balance + a.CreditLimit - a.Held
}

func (a *Account) Outstanding() int64 {
	if a.Balance < 0 {
		return -a.Balance
	}
	return 0
}

func (a *Account) SetCreditLimit(limit int64, p Program, now time.Time) error {
	if a.Mode != ModeCredit {
		return Invalid("credit limit only applies to CREDIT accounts")
	}
	if a.Status == AccountCancelled {
		return Conflict("account is CANCELLED")
	}
	if limit <= 0 || limit > p.MaxCreditLimit {
		return Invalid("credit_limit must be between 1 and %d", p.MaxCreditLimit)
	}
	if limit < a.Outstanding()+a.Held {
		return Conflict("credit limit %d is below the amount already used %d", limit, a.Outstanding()+a.Held)
	}
	a.CreditLimit = limit
	a.UpdatedAt = now
	return nil
}

func (a *Account) Block(reason string, now time.Time) error {
	if a.Status != AccountNormal {
		return Conflict("account is %s, only a NORMAL account can be blocked", a.Status)
	}
	a.Status, a.StatusReason, a.UpdatedAt = AccountBlocked, strings.TrimSpace(reason), now
	return nil
}

func (a *Account) Unblock(now time.Time) error {
	if a.Status != AccountBlocked {
		return Conflict("account is %s, only a BLOCKED account can be unblocked", a.Status)
	}
	a.Status, a.StatusReason, a.UpdatedAt = AccountNormal, "", now
	return nil
}

func (a *Account) Cancel(reason string, now time.Time) error {
	if a.Status == AccountCancelled {
		return Conflict("account is already CANCELLED")
	}
	if a.Held > 0 {
		return Conflict("account has %d in pending authorizations", a.Held)
	}
	if a.Balance < 0 {
		return Conflict("account has an outstanding balance of %d", -a.Balance)
	}
	a.Status, a.StatusReason, a.UpdatedAt = AccountCancelled, strings.TrimSpace(reason), now
	return nil
}

func (a *Account) Hold(amount int64, now time.Time) {
	a.Held += amount
	a.UpdatedAt = now
}

func (a *Account) Release(amount int64, now time.Time) {
	a.Held -= amount
	if a.Held < 0 {
		a.Held = 0
	}
	a.UpdatedAt = now
}

func (a *Account) Post(amount int64, now time.Time) {
	a.Balance += amount
	a.UpdatedAt = now
}

func (a *Account) ReceivePayment(amount int64, now time.Time) error {
	if a.Status == AccountCancelled {
		return Conflict("account is CANCELLED")
	}
	if amount <= 0 || amount > maxAmount {
		return Invalid("amount must be between 1 and %d", int64(maxAmount))
	}
	a.Post(amount, now)
	return nil
}

type TransactionType string

const (
	TxnPurchase   TransactionType = "PURCHASE"
	TxnWithdrawal TransactionType = "WITHDRAWAL"
	TxnRefund     TransactionType = "REFUND"
	TxnPayment    TransactionType = "PAYMENT"
)

type Transaction struct {
	ID              string
	AccountID       string
	CardID          string
	AuthorizationID string
	Type            TransactionType
	ProcessingCode  ProcessingCode
	Amount          int64
	BalanceAfter    int64
	Description     string
	CreatedAt       time.Time
}
