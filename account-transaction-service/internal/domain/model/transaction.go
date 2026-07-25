package model

import "time"

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypePayment    TransactionType = "PAYMENT"
)

// Transaction is a single ledger entry. Only the fields relevant to Type
// are populated: AccountID for deposit/withdrawal/payment, SenderID/ReceiverID
// for transfer, ServiceName for payment.
type Transaction struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	Type        TransactionType `json:"type"`
	Amount      float64         `json:"amount"`
	DateCreated time.Time       `json:"date_created"`

	AccountID   string `json:"account_id,omitempty"`
	SenderID    string `json:"sender_id,omitempty"`
	ReceiverID  string `json:"receiver_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}

func (Transaction) TableName() string { return "transactions" }
