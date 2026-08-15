package wallet

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type OwnerType string

const (
	OwnerTypeUser      OwnerType = "user"
	OwnerTypeCorporate OwnerType = "corporate"
)

type Wallet struct {
	ID               shared.WalletID
	OwnerType        OwnerType
	OwnerID          string
	Balance          shared.Money
	Version          int
	PersistedVersion int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewWallet(ownerType OwnerType, ownerID, currency string, now time.Time) (*Wallet, error) {
	if ownerID == "" {
		return nil, ErrInvalidAmount
	}
	return &Wallet{
		ID:        shared.NewWalletID(),
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Balance:   shared.ZeroMoney(currency),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (w *Wallet) Credit(amount shared.Money, now time.Time) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	newBalance, err := w.Balance.Add(amount)
	if err != nil {
		return err
	}
	w.Balance = newBalance
	w.Version++
	w.UpdatedAt = now
	return nil
}

func (w *Wallet) Debit(amount shared.Money, now time.Time) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if w.Balance.LessThan(amount) {
		return ErrInsufficientFunds
	}
	newBalance, err := w.Balance.Sub(amount)
	if err != nil {
		return err
	}
	w.Balance = newBalance
	w.Version++
	w.UpdatedAt = now
	return nil
}
