package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyVND Currency = "VND"
)

type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "ACTIVE"
	WalletStatusFrozen    WalletStatus = "FROZEN"
	WalletStatusClosed    WalletStatus = "CLOSED"
	WalletStatusSuspended WalletStatus = "SUSPENDED"
)

type Wallet struct {
	ID           uuid.UUID       `json:"id"            db:"id"`
	UserID       uuid.UUID       `json:"user_id"       db:"user_id"`
	Balance      decimal.Decimal `json:"balance"       db:"balance"`
	FrozenAmount decimal.Decimal `json:"frozen_amount" db:"frozen_amount"`
	Currency     Currency        `json:"currency"      db:"currency"`
	Status       WalletStatus    `json:"status"        db:"status"`
	Version      int             `json:"version"       db:"version"` 
	CreatedAt    time.Time       `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"    db:"updated_at"`
}

func (w *Wallet) AvailableBalance() decimal.Decimal {
	return w.Balance.Sub(w.FrozenAmount)
}

func (w *Wallet) FreezeAmount(amount decimal.Decimal) error {
	if w.AvailableBalance().LessThan(amount) {
		return ErrInsufficientFunds
	}
	w.FrozenAmount = w.FrozenAmount.Add(amount)
	w.UpdatedAt = time.Now()
	return nil
}

func (w *Wallet) UnfreezeAmount(amount decimal.Decimal) {
	w.FrozenAmount = w.FrozenAmount.Sub(amount)
	if w.FrozenAmount.IsNegative() {
		w.FrozenAmount = decimal.Zero
	}
	w.UpdatedAt = time.Now()
}

func (w *Wallet) Debit(amount decimal.Decimal) error {
	if w.FrozenAmount.LessThan(amount) {
		return ErrInsufficientFunds
	}
	w.Balance = w.Balance.Sub(amount)
	w.FrozenAmount = w.FrozenAmount.Sub(amount)
	w.UpdatedAt = time.Now()
	return nil
}

func (w *Wallet) Credit(amount decimal.Decimal) error {
	if amount.IsNegative() || amount.IsZero() {
		return ErrInsufficientFunds
	}
	w.Balance = w.Balance.Add(amount)
	w.UpdatedAt = time.Now()
	return nil
}
