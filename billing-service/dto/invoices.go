package dto

type Invoices struct {
	InvoiceID      int
	SubscriptionID int
	InvoiceDate    int
	DueDate        int
	Amount         float32
	Tax            string
	Status         string
}
