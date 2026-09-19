package model

import "time"

type PaymentMethod string

const (
	Cash    PaymentMethod = "CASH"
	Card    PaymentMethod = "CARD"
	EWallet PaymentMethod = "E_WALLET"
)

func (m PaymentMethod) Valid() bool {
	switch m {
	case Cash, Card, EWallet:
		return true
	default:
		return false
	}
}

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "PENDING"
	PaymentPaid    PaymentStatus = "PAID"
	PaymentFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	PaymentID string        `gorm:"column:payment_id;primaryKey"`
	TicketID  string        `gorm:"column:history_id"`
	Amount    float64       `gorm:"column:amount"`
	Method    PaymentMethod `gorm:"column:method"`
	PaidAt    *time.Time    `gorm:"column:paid_at"`
	Status    PaymentStatus `gorm:"column:status"`
}

func (Payment) TableName() string { return "payments" }
