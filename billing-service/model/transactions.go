package model

type Transactions struct {
	TransactionID   int
	InvoiceID       int
	PaymentMethod   string
	TransactionDate int
	Amount          float64
	Status          string
}
