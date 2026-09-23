package model

import "time"

type Balance struct {
	UserID    string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	Available int64     `gorm:"column:available" json:"available"`
	Locked    int64     `gorm:"column:locked" json:"locked"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Balance) TableName() string { return "balances" }

type LedgerType string

const (
	LedgerDeposit        LedgerType = "deposit"
	LedgerWithdraw       LedgerType = "withdraw"
	LedgerWithdrawRevert LedgerType = "withdraw_reversal"
	LedgerSettlement     LedgerType = "settlement"
	LedgerConversion     LedgerType = "conversion"
	LedgerBond           LedgerType = "dispute_bond"
	LedgerBondRefund     LedgerType = "dispute_bond_refund"
	LedgerFunding        LedgerType = "exchange_funding"
)

type LedgerEntry struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     string     `gorm:"column:user_id" json:"user_id"`
	Type       LedgerType `gorm:"column:type" json:"type"`
	Amount     int64      `gorm:"column:amount" json:"amount"` // signed: negative leaves the balance
	MarketID   int64      `gorm:"column:market_id" json:"market_id,omitempty"`
	TransferID string     `gorm:"column:transfer_id" json:"transfer_id,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
}

func (LedgerEntry) TableName() string { return "ledger_entries" }

type Position struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MarketID     int64     `gorm:"column:market_id" json:"market_id"`
	UserID       string    `gorm:"column:user_id" json:"user_id"`
	Outcome      Outcome   `gorm:"column:outcome" json:"outcome"`
	Shares       int64     `gorm:"column:shares" json:"shares"`
	LockedShares int64     `gorm:"column:locked_shares" json:"locked_shares"`
	CostBasis    int64     `gorm:"column:cost_basis" json:"cost_basis"`
	RealizedPnL  int64     `gorm:"column:realized_pnl" json:"realized_pnl"`
	Settled      bool      `gorm:"column:settled" json:"settled"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Position) TableName() string { return "positions" }

func (p *Position) Free() int64 { return p.Shares - p.LockedShares }
