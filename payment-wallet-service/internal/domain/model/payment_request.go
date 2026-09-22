package model

import "time"

type PaymentRequestStatus string

const (
	PaymentRequestPending   PaymentRequestStatus = "PENDING"
	PaymentRequestPaid      PaymentRequestStatus = "PAID"
	PaymentRequestCancelled PaymentRequestStatus = "CANCELLED"
	PaymentRequestExpired   PaymentRequestStatus = "EXPIRED"
)

type PaymentRequest struct {
	PaymentRequestID  string               `gorm:"column:payment_request_id;primaryKey" json:"payment_request_id"`
	RequesterWalletID string               `gorm:"column:requester_wallet_id" json:"requester_wallet_id"`
	PayerWalletID     *string              `gorm:"column:payer_wallet_id" json:"payer_wallet_id,omitempty"`
	Amount            int64                `gorm:"column:amount" json:"amount"`
	Currency          string               `gorm:"column:currency" json:"currency"`
	Status            PaymentRequestStatus `gorm:"column:status" json:"status"`
	Description       string               `gorm:"column:description" json:"description,omitempty"`
	TransferID        *string              `gorm:"column:transfer_id" json:"transfer_id,omitempty"`
	ExpiresAt         *time.Time           `gorm:"column:expires_at" json:"expires_at,omitempty"`
	CreatedAt         time.Time            `gorm:"column:created_at" json:"created_at"`
}

func (PaymentRequest) TableName() string { return "payment_requests" }

func (r PaymentRequest) Expired(now time.Time) bool {
	return r.ExpiresAt != nil && now.After(*r.ExpiresAt)
}
