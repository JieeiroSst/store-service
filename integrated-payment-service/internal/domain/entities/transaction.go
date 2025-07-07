package entities

import "time"

type Transaction struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	PaymentID string    `json:"payment_id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Payment   Payment   `json:"payment" gorm:"foreignKey:PaymentID"`
}
