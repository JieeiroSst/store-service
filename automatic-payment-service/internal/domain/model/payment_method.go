package model

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod struct {
	PaymentMethodID uuid.UUID `gorm:"column:payment_method_id;primaryKey" json:"paymentMethodId"`
	UserID          uuid.UUID `gorm:"column:user_id;index;not null" json:"userId"`
	Provider        string    `gorm:"column:provider;not null" json:"provider"` // visa, mastercard, paypal, ...
	TokenID         string    `gorm:"column:token_id" json:"tokenId,omitempty"`
	LastFourDigits  string    `gorm:"column:last_four_digits" json:"lastFourDigits,omitempty"`
	ExpiryDate      string    `gorm:"column:expiry_date" json:"expiryDate,omitempty"` // MM/YYYY
	IsDefault       bool      `gorm:"column:is_default;not null;default:false" json:"isDefault"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (PaymentMethod) TableName() string { return "payment_methods" }
