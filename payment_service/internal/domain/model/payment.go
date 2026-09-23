package model

import "time"

type Provider string

const (
	ProviderPayPal   Provider = "paypal"
	ProviderPayoneer Provider = "payoneer"
	ProviderStripe   Provider = "stripe"
	ProviderWise     Provider = "wise"
)

func (p Provider) Valid() bool {
	switch p {
	case ProviderPayPal, ProviderPayoneer, ProviderStripe, ProviderWise:
		return true
	default:
		return false
	}
}

type PaymentStatus string

const (
	PaymentStatusPending           PaymentStatus = "pending"
	PaymentStatusAuthorized        PaymentStatus = "authorized"
	PaymentStatusCaptured          PaymentStatus = "captured"
	PaymentStatusFailed            PaymentStatus = "failed"
	PaymentStatusRefunded          PaymentStatus = "refunded"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
)

type Payment struct {
	ID             int64 `gorm:"primaryKey"`
	Provider       Provider
	ExternalID     string
	Amount         int64
	Currency       string
	Status         PaymentStatus
	PayerEmail     string
	Description    string
	FailureReason  string
	IdempotencyKey string
	RefundedAmount int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Payment) TableName() string {
	return "payments"
}
