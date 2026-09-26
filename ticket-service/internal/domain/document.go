package domain

import "time"

const (
	DocInvoice = "invoice" // the receipt
	DocTickets = "tickets" // the e-tickets
)

func ValidDocKind(k string) bool { return k == DocInvoice || k == DocTickets }

type OrderDocument struct {
	OrderID   int64
	Kind      string
	FileID    string
	FileName  string
	Size      int64
	CreatedAt time.Time
}

type TicketPage struct {
	TicketID   int64
	TypeName   string
	Seat       string
	HolderName string
	Code       string
	Status     string
}

type TicketDocument struct {
	OrderID    int64
	EventTitle string
	StartsAt   time.Time
	EndsAt     time.Time
	Venue      string
	Address    string
	BuyerName  string
	Location   *time.Location
	Tickets    []TicketPage
}
