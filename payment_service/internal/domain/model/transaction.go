package model

import "time"

type TransactionType string

const (
	TransactionTypeCreate  TransactionType = "create"
	TransactionTypeRefund  TransactionType = "refund"
	TransactionTypeWebhook TransactionType = "webhook"
)

type Transaction struct {
	ID           int64 `gorm:"primaryKey"`
	PaymentID    int64
	Type         TransactionType
	Status       PaymentStatus
	Amount       int64
	ExternalID   string
	RawStatus    string
	ErrorMessage string
	CreatedAt    time.Time
}

func (Transaction) TableName() string {
	return "transactions"
}
