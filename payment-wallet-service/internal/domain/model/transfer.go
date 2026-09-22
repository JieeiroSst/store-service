package model

import "time"

type Transfer struct {
	TransferID       string            `gorm:"column:transfer_id;primaryKey" json:"transfer_id"`
	SenderWalletID   string            `gorm:"column:sender_wallet_id" json:"sender_wallet_id"`
	ReceiverWalletID string            `gorm:"column:receiver_wallet_id" json:"receiver_wallet_id"`
	Amount           int64             `gorm:"column:amount" json:"amount"`
	Currency         string            `gorm:"column:currency" json:"currency"`
	Status           TransactionStatus `gorm:"column:status" json:"status"`
	ReferenceID      string            `gorm:"column:reference_id" json:"reference_id,omitempty"`
	OutTransactionID string            `gorm:"column:out_transaction_id" json:"out_transaction_id"`
	InTransactionID  string            `gorm:"column:in_transaction_id" json:"in_transaction_id"`
	Description      string            `gorm:"column:description" json:"description,omitempty"`
	CreatedAt        time.Time         `gorm:"column:created_at" json:"created_at"`
}

func (Transfer) TableName() string { return "transfers" }
