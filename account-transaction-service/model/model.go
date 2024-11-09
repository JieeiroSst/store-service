package model

import "time"

type Account struct {
	ID           int           `json:"id"`
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	CreatedAt    time.Time     `json:"created_at"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	Account   *Account  `json:"account"`
}

type AccountToAccountTransaction struct {
	ID         int `json:"id"`
	SenderID   int `json:"sender_id"`
	ReceiverID int `json:"receiver_id"`
}

type WithdrawnTransaction struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type DepositTransaction struct {
	ID        int `json:"id"`
	AccountID int `json:"account_id"`
}

type PaymentForServiceTransaction struct {
	ID          int    `json:"id"`
	AccountID   int    `json:"account_id"`
	ServiceName string `json:"service_name"`
}
