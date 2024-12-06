package model

import "time"

// Payers: Individuals or entities making payments.
type Payers struct {
	PayerID     int    `json:"payer_id" cql:"payer_id"`
	Name        string `json:"name" cql:"name"`
	Email       string `json:"email" cql:"email"`
	PhoneNumber string `json:"phone_number" cql:"phone_number"`
}

// Buyers: Individuals or entities receiving payments.
type Buyers struct {
	BuyerID     int    `json:"buyer_id" cql:"buyer_id"`
	Name        string `json:"name" cql:"name"`
	Email       string `json:"email" cql:"email"`
	PhoneNumber string `json:"phone_number" cql:"phone_number"`
}

type Transactions struct {
	TransactionID   int       `json:"transaction_id" cql:"transaction_id"`
	PayerID         int       `json:"payer_id" cql:"payer_id"`
	BuyerID         int       `json:"buyer_id" cql:"buyer_id"`
	Amount          float64   `json:"amount" cql:"amount"`
	TransactionDate time.Time `json:"transaction_date" cql:"transaction_date"`
	TransactionType int       `json:"transaction_type" cql:"transaction_type"`
	Description     string    `json:"description" cql:"description"`
	Status          int       `json:"status" cql:"status"`
}
