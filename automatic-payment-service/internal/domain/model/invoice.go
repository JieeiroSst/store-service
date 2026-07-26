package model

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoicePaid   InvoiceStatus = "paid"
	InvoiceUnpaid InvoiceStatus = "unpaid"
	InvoiceVoid   InvoiceStatus = "void"
)

type Invoice struct {
	InvoiceID      uuid.UUID     `gorm:"column:invoice_id;primaryKey" json:"invoiceId"`
	TransactionID  uuid.UUID     `gorm:"column:transaction_id;not null" json:"transactionId"`
	SubscriptionID uuid.UUID     `gorm:"column:subscription_id;index;not null" json:"subscriptionId"`
	UserID         uuid.UUID     `gorm:"column:user_id;index;not null" json:"userId"`
	Amount         float64       `gorm:"column:amount;not null" json:"amount"`
	TaxAmount      float64       `gorm:"column:tax_amount;not null;default:0" json:"taxAmount"`
	TotalAmount    float64       `gorm:"column:total_amount;not null" json:"totalAmount"`
	Status         InvoiceStatus `gorm:"column:status;not null" json:"status"`
	DueDate        time.Time     `gorm:"column:due_date;not null" json:"dueDate"`
	PaidDate       *time.Time    `gorm:"column:paid_date" json:"paidDate,omitempty"`
	InvoiceNumber  string        `gorm:"column:invoice_number;uniqueIndex;not null" json:"invoiceNumber"`
	CreatedAt      time.Time     `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Invoice) TableName() string { return "invoices" }
