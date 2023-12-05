package entity

type Transaction struct {
	TransactionID   int    `json:"transaction_id" gorm:"column:transaction_id"`
	TransactionType string `json:"transaction_type" gorm:"column:transaction_type"`
	Amount          uint   `json:"amount" gorm:"column:amount"`
	TransactionDate int    `json:"transaction_date" gorm:"column:transaction_date"`
}

func (Transaction) TableName() string {
	return "transactions"
}