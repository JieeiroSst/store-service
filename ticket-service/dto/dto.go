package dto

import "time"

type Customers struct {
	CustomerID   int    `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Address      string `json:"address"`
}

type Tickets struct {
	TicketID    int       `json:"ticket_id"`
	TicketName  string    `json:"ticket_name"`
	StartDate   time.Time `json:"start_date"`
	AddressRoom string    `json:"address_room"`
	Amount      float64   `json:"amount"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"`
}

type Invoices struct {
	InvoicesID  int       `json:"invoices_id"`
	CustomerID  int       `json:"customer_id"`
	BuyDate     time.Time `json:"buy_date"`
	TotalAmount float64   `json:"total_amount"`
	Note        string    `json:"note"`
	Customers   Customers `gorm:"references:CustomerID"`
}

type InvoiceDetails struct {
	InvoiceDetailID int      `json:"invoice_detail_id"`
	InvoicesID      int      `json:"invoices_id"`
	TicketID        int      `json:"ticket_id"`
	Quantity        int      `json:"quantity"`
	Status          int      `json:"status"`
	Invoices        Invoices `gorm:"references:InvoicesID"`
	Tickets         Tickets  `gorm:"references:TicketID"`
}

type CreateTicketsRequest struct {
	TicketID    int       `json:"ticket_id"`
	TicketName  string    `json:"ticket_name"`
	StartDate   time.Time `json:"start_date"`
	AddressRoom string    `json:"address_room"`
	Amount      float64   `json:"amount"`
	Quantity    int       `json:"quantity"`
}

type BuyTicketRequest struct {
	CustomerID int `json:"customer_id"`
	TicketID   int `json:"ticket_id"`
	Quantity   int `json:"quantity"`
}
