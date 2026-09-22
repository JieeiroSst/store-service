package model

import "time"

type PaymentMethodType string

const (
	BankAccount PaymentMethodType = "BANK_ACCOUNT"
	Card        PaymentMethodType = "CARD"
)

func (t PaymentMethodType) Valid() bool {
	switch t {
	case BankAccount, Card:
		return true
	default:
		return false
	}
}

type PaymentMethod struct {
	PaymentMethodID string            `gorm:"column:payment_method_id;primaryKey" json:"payment_method_id"`
	UserID          string            `gorm:"column:user_id" json:"user_id"`
	Type            PaymentMethodType `gorm:"column:type" json:"type"`
	Provider        string            `gorm:"column:provider" json:"provider"`
	AccountNumber   string            `gorm:"column:account_number" json:"account_number"`
	IsDefault       bool              `gorm:"column:is_default" json:"is_default"`
	IsActive        bool              `gorm:"column:is_active" json:"is_active"`
	CreatedAt       time.Time         `gorm:"column:created_at" json:"created_at"`
}

func (PaymentMethod) TableName() string { return "payment_methods" }
