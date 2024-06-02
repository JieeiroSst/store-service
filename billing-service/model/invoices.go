package model

type Invoices struct {
	InvoiceID      int
	SubscriptionID int
	InvoiceDate    int
	DueDate        int
	Amount         float64
	Tax            string
	Status         string
}
