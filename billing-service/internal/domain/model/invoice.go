package model

type Invoice struct {
	InvoiceID      int     `gorm:"column:invoice_id;primaryKey;autoIncrement" json:"invoice_id,omitempty"`
	SubscriptionID int     `gorm:"column:subscription_id" json:"subscription_id"`
	InvoiceDate    int64   `gorm:"column:invoice_date" json:"invoice_date"`
	DueDate        int64   `gorm:"column:due_date" json:"due_date"`
	Amount         float64 `gorm:"column:amount" json:"amount"`
	Tax            string  `gorm:"column:tax" json:"tax,omitempty"`
	Status         string  `gorm:"column:status" json:"status"`
}

func (Invoice) TableName() string { return "invoices" }
