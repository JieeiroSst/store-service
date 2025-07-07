package entities

import (
	"time"
)

type PaymentStatus string
type PaymentMethod string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSuccess    PaymentStatus = "success"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCanceled   PaymentStatus = "canceled"
	PaymentStatusUnknown    PaymentStatus = "unknown"
)

const (
	PaymentMethodMomo         PaymentMethod = "momo"
	PaymentMethodVNPay        PaymentMethod = "vnpay"
	PaymentMethodZaloPay      PaymentMethod = "zalopay"
	PaymentMethodPayPal       PaymentMethod = "paypal"
	PaymentMethodStripe       PaymentMethod = "stripe"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

type Payment struct {
	ID                string        `json:"id" gorm:"primaryKey"`
	UserID            string        `json:"user_id"`
	Amount            float64       `json:"amount"`
	Currency          string        `json:"currency"`
	PaymentMethod     PaymentMethod `json:"payment_method"`
	Status            PaymentStatus `json:"status"`
	Description       string        `json:"description"`
	Metadata          string        `json:"metadata"`
	ExternalID        string        `json:"external_id"`
	ProcessorResponse string        `json:"processor_response"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	User              User          `json:"user" gorm:"foreignKey:UserID"`
}
