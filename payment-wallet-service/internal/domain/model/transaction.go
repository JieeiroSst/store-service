package model

import "time"

type TransactionType string

const (
	TxnDeposit     TransactionType = "DEPOSIT"
	TxnWithdraw    TransactionType = "WITHDRAW"
	TxnTransferIn  TransactionType = "TRANSFER_IN"
	TxnTransferOut TransactionType = "TRANSFER_OUT"
	TxnReversal    TransactionType = "REVERSAL"
	TxnPocketOut   TransactionType = "POCKET_OUT"
	TxnPocketIn    TransactionType = "POCKET_IN"
)

type TransactionStatus string

const (
	TxnPending   TransactionStatus = "PENDING"
	TxnCompleted TransactionStatus = "COMPLETED"
	TxnFailed    TransactionStatus = "FAILED"
	TxnReversed  TransactionStatus = "REVERSED"
)

type Transaction struct {
	TransactionID        string            `gorm:"column:transaction_id;primaryKey" json:"transaction_id"`
	WalletID             string            `gorm:"column:wallet_id" json:"wallet_id"`
	Type                 TransactionType   `gorm:"column:type" json:"type"`
	Amount               int64             `gorm:"column:amount" json:"amount"`
	Currency             string            `gorm:"column:currency" json:"currency"`
	Status               TransactionStatus `gorm:"column:status" json:"status"`
	ReferenceID          string            `gorm:"column:reference_id" json:"reference_id,omitempty"`
	CounterpartyWalletID *string           `gorm:"column:counterparty_wallet_id" json:"counterparty_wallet_id,omitempty"`
	PocketID             *string           `gorm:"column:pocket_id" json:"pocket_id,omitempty"`
	Description          string            `gorm:"column:description" json:"description,omitempty"`
	CreatedAt            time.Time         `gorm:"column:created_at" json:"created_at"`
}

func (Transaction) TableName() string { return "transactions" }
