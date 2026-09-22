package model

import "time"

type WalletStatus string

const (
	WalletActive WalletStatus = "ACTIVE"
	WalletFrozen WalletStatus = "FROZEN"
	WalletClosed WalletStatus = "CLOSED"
)

type Wallet struct {
	WalletID            string       `gorm:"column:wallet_id;primaryKey" json:"wallet_id"`
	UserID              string       `gorm:"column:user_id" json:"user_id"`
	Balance             int64        `gorm:"column:balance" json:"balance"`
	Currency            string       `gorm:"column:currency" json:"currency"`
	Status              WalletStatus `gorm:"column:status" json:"status"`
	DailyLimit          int64        `gorm:"column:daily_limit" json:"daily_limit"`
	PerTransactionLimit int64        `gorm:"column:per_transaction_limit" json:"per_transaction_limit"`
	FrozenReason        *string      `gorm:"column:frozen_reason" json:"frozen_reason,omitempty"`
	CreatedAt           time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time    `gorm:"column:updated_at" json:"updated_at"`
}

func (Wallet) TableName() string { return "wallets" }
