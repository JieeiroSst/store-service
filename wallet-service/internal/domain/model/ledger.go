package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EntryDirection string

const (
	EntryDebit  EntryDirection = "DEBIT"
	EntryCredit EntryDirection = "CREDIT"
)

type LedgerEntry struct {
	ID            uuid.UUID       `json:"id"             db:"id"`
	WalletID      uuid.UUID       `json:"wallet_id"      db:"wallet_id"`
	TransactionID uuid.UUID       `json:"transaction_id" db:"transaction_id"`
	Direction     EntryDirection  `json:"direction"      db:"direction"`
	Amount        decimal.Decimal `json:"amount"         db:"amount"`
	BalanceAfter  decimal.Decimal `json:"balance_after"  db:"balance_after"`
	CreatedAt     time.Time       `json:"created_at"     db:"created_at"`
}

func NewLedgerEntry(w *Wallet, txID uuid.UUID, direction EntryDirection, amount decimal.Decimal, now time.Time) *LedgerEntry {
	return &LedgerEntry{
		ID:            uuid.New(),
		WalletID:      w.ID,
		TransactionID: txID,
		Direction:     direction,
		Amount:        amount,
		BalanceAfter:  w.Balance,
		CreatedAt:     now,
	}
}

type LedgerSummary struct {
	OpeningBalance decimal.Decimal
	ClosingBalance decimal.Decimal
	TotalCredit    decimal.Decimal
	TotalDebit     decimal.Decimal
	CreditCount    int64
	DebitCount     int64
}

type Statement struct {
	WalletID       uuid.UUID       `json:"wallet_id"`
	Currency       Currency        `json:"currency"`
	From           time.Time       `json:"from"`
	To             time.Time       `json:"to"`
	OpeningBalance decimal.Decimal `json:"opening_balance"`
	ClosingBalance decimal.Decimal `json:"closing_balance"`
	TotalCredit    decimal.Decimal `json:"total_credit"`
	TotalDebit     decimal.Decimal `json:"total_debit"`
	NetChange      decimal.Decimal `json:"net_change"`
	CreditCount    int64           `json:"credit_count"`
	DebitCount     int64           `json:"debit_count"`
}

func NewStatement(w *Wallet, from, to time.Time, s LedgerSummary) Statement {
	return Statement{
		WalletID:       w.ID,
		Currency:       w.Currency,
		From:           from,
		To:             to,
		OpeningBalance: s.OpeningBalance,
		ClosingBalance: s.ClosingBalance,
		TotalCredit:    s.TotalCredit,
		TotalDebit:     s.TotalDebit,
		NetChange:      s.TotalCredit.Sub(s.TotalDebit),
		CreditCount:    s.CreditCount,
		DebitCount:     s.DebitCount,
	}
}
