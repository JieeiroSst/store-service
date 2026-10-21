package model

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "ACTIVE"
	WalletStatusFrozen    WalletStatus = "FROZEN"
	WalletStatusClosed    WalletStatus = "CLOSED"
	WalletStatusSuspended WalletStatus = "SUSPENDED"
)

type Tier string

const (
	TierBasic    Tier = "BASIC"
	TierVerified Tier = "VERIFIED"
	TierPremium  Tier = "PREMIUM"
)

func (t Tier) Valid() bool {
	_, ok := tierLimitsUSD[t]
	return ok
}

const (
	PinMaxAttempts  = 5
	PinLockDuration = 30 * time.Minute
)

var pinPattern = regexp.MustCompile(`^[0-9]{6}$`)

func ValidatePinFormat(pin string) error {
	if !pinPattern.MatchString(pin) {
		return ErrPinFormat
	}
	return nil
}

type Wallet struct {
	ID                uuid.UUID       `json:"id"            db:"id"`
	UserID            uuid.UUID       `json:"user_id"       db:"user_id"`
	Balance           decimal.Decimal `json:"balance"       db:"balance"`
	FrozenAmount      decimal.Decimal `json:"frozen_amount" db:"frozen_amount"`
	Currency          Currency        `json:"currency"      db:"currency"`
	Status            WalletStatus    `json:"status"        db:"status"`
	Tier              Tier            `json:"tier"          db:"tier"`
	HasPin            bool            `json:"has_pin"       db:"has_pin"`
	PinHash           string          `json:"-"             db:"pin_hash"`
	PinFailedAttempts int             `json:"-"             db:"pin_failed_attempts"`
	PinLockedUntil    *time.Time      `json:"-"             db:"pin_locked_until"`
	Version           int             `json:"version"       db:"version"`
	CreatedAt         time.Time       `json:"created_at"    db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"    db:"updated_at"`
}

type Balance struct {
	WalletID  uuid.UUID       `json:"wallet_id"`
	Currency  Currency        `json:"currency"`
	Balance   decimal.Decimal `json:"balance"`
	Held      decimal.Decimal `json:"held"`
	Available decimal.Decimal `json:"available"`
	Status    WalletStatus    `json:"status"`
}

func NewWallet(userID uuid.UUID, currency Currency, now time.Time) (*Wallet, error) {
	if userID == uuid.Nil {
		return nil, Invalid("user_id is required")
	}
	if !currency.Valid() {
		return nil, ErrUnsupportedCurrency
	}
	return &Wallet{
		ID:           uuid.New(),
		UserID:       userID,
		Balance:      decimal.Zero,
		FrozenAmount: decimal.Zero,
		Currency:     currency,
		Status:       WalletStatusActive,
		Tier:         TierBasic,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (w *Wallet) AvailableBalance() decimal.Decimal {
	return w.Balance.Sub(w.FrozenAmount)
}

func (w *Wallet) Snapshot() Balance {
	return Balance{
		WalletID:  w.ID,
		Currency:  w.Currency,
		Balance:   w.Balance,
		Held:      w.FrozenAmount,
		Available: w.AvailableBalance(),
		Status:    w.Status,
	}
}

func (w *Wallet) EnsureOwner(userID uuid.UUID) error {
	if w.UserID != userID {
		return ErrNotWalletOwner
	}
	return nil
}

func (w *Wallet) EnsureActive() error {
	switch w.Status {
	case WalletStatusActive:
		return nil
	case WalletStatusClosed:
		return ErrWalletClosed
	}
	return ErrWalletInactive
}

func (w *Wallet) Hold(amount decimal.Decimal, now time.Time) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if w.AvailableBalance().LessThan(amount) {
		return ErrInsufficientFunds
	}
	w.FrozenAmount = w.FrozenAmount.Add(amount)
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) ReleaseHold(amount decimal.Decimal, now time.Time) {
	w.FrozenAmount = w.FrozenAmount.Sub(amount)
	if w.FrozenAmount.IsNegative() {
		w.FrozenAmount = decimal.Zero
	}
	w.UpdatedAt = now
}

func (w *Wallet) CaptureHold(amount decimal.Decimal, now time.Time) error {
	if w.FrozenAmount.LessThan(amount) || w.Balance.LessThan(amount) {
		return ErrInsufficientFunds
	}
	w.Balance = w.Balance.Sub(amount)
	w.FrozenAmount = w.FrozenAmount.Sub(amount)
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) Debit(amount decimal.Decimal, now time.Time) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if w.AvailableBalance().LessThan(amount) {
		return ErrInsufficientFunds
	}
	w.Balance = w.Balance.Sub(amount)
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) Credit(amount decimal.Decimal, now time.Time) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if w.Status == WalletStatusClosed {
		return ErrWalletClosed
	}
	w.Balance = w.Balance.Add(amount)
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) Freeze(now time.Time) error {
	if w.Status == WalletStatusClosed {
		return ErrWalletClosed
	}
	w.Status = WalletStatusFrozen
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) Unfreeze(now time.Time) error {
	if w.Status == WalletStatusClosed {
		return ErrWalletClosed
	}
	w.Status = WalletStatusActive
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) Close(now time.Time) error {
	if w.Status == WalletStatusClosed {
		return nil
	}
	if !w.Balance.IsZero() || !w.FrozenAmount.IsZero() {
		return ErrWalletNotEmpty
	}
	w.Status = WalletStatusClosed
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) ChangeTier(tier Tier, now time.Time) error {
	if !tier.Valid() {
		return ErrInvalidTier
	}
	w.Tier = tier
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) SetPinHash(hash string, now time.Time) {
	w.PinHash = hash
	w.HasPin = hash != ""
	w.PinFailedAttempts = 0
	w.PinLockedUntil = nil
	w.UpdatedAt = now
}

func (w *Wallet) PinLocked(now time.Time) bool {
	return w.PinLockedUntil != nil && now.Before(*w.PinLockedUntil)
}

func (w *Wallet) RegisterPinFailure(now time.Time) bool {
	w.PinFailedAttempts++
	w.UpdatedAt = now
	if w.PinFailedAttempts >= PinMaxAttempts {
		until := now.Add(PinLockDuration)
		w.PinLockedUntil = &until
		w.PinFailedAttempts = 0
		return true
	}
	return false
}

func (w *Wallet) ResetPinFailures(now time.Time) bool {
	if w.PinFailedAttempts == 0 && w.PinLockedUntil == nil {
		return false
	}
	w.PinFailedAttempts = 0
	w.PinLockedUntil = nil
	w.UpdatedAt = now
	return true
}
