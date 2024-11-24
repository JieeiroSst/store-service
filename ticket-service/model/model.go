package model

import "time"

type Customers struct {
	CustomerID   int
	CustomerName string
	Phone        string
	Email        string
	Address      string
}

type Tickets struct {
	TicketID    int
	TicketName  string
	StartDate   time.Time
	AddressRoom string
	Amount      float64
	Quantity    int
	Status      string
}

type Invoices struct {
	InvoicesID  int
	CustomerID  int
	BuyDate     time.Time
	TotalAmount float64
	Note        string
	Customers   Customers
}

type InvoiceDetails struct {
	InvoiceDetailID int
	InvoicesID      int
	TicketID        int
	Quantity        int
	Invoices        Invoices
	Tickets         Tickets
}
