package model

type Transaction struct {
	TransactionID   int     `gorm:"column:transaction_id;primaryKey;autoIncrement" json:"transaction_id,omitempty"`
	InvoiceID       int     `gorm:"column:invoice_id" json:"invoice_id"`
	PaymentMethod   string  `gorm:"column:payment_method" json:"payment_method"`
	TransactionDate int64   `gorm:"column:transaction_date" json:"transaction_date"`
	Amount          float64 `gorm:"column:amount" json:"amount"`
	Status          string  `gorm:"column:status" json:"status"`
}

func (Transaction) TableName() string { return "transactions" }
