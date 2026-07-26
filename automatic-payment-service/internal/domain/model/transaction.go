package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	TransactionSuccessful TransactionStatus = "successful"
	TransactionFailed     TransactionStatus = "failed"
	TransactionPending    TransactionStatus = "pending"
	TransactionRefunded   TransactionStatus = "refunded"
)

type Transaction struct {
	TransactionID        uuid.UUID         `gorm:"column:transaction_id;primaryKey" json:"transactionId"`
	SubscriptionID       uuid.UUID         `gorm:"column:subscription_id;index;not null" json:"subscriptionId"`
	PaymentMethodID      uuid.UUID         `gorm:"column:payment_method_id" json:"paymentMethodId,omitempty"`
	Amount               float64           `gorm:"column:amount;not null" json:"amount"`
	Currency             string            `gorm:"column:currency;not null;default:USD" json:"currency"`
	Status               TransactionStatus `gorm:"column:status;not null" json:"status"`
	GatewayTransactionID string            `gorm:"column:gateway_transaction_id" json:"gatewayTransactionId,omitempty"`
	ErrorMessage         string            `gorm:"column:error_message" json:"errorMessage,omitempty"`
	CreatedAt            time.Time         `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Transaction) TableName() string { return "transactions" }
