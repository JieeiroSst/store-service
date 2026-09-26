package domain

import (
	"fmt"
	"time"
)

type Seller struct {
	Name    string
	TaxID   string
	Address string
}

type InvoiceLine struct {
	Description string
	Quantity    int
	UnitPrice   int64
	Amount      int64
}

type Invoice struct {
	Number       string
	IssuedAt     time.Time
	OrderID      int64
	Seller       Seller
	BuyerName    string
	BuyerEmail   string
	EventTitle   string
	EventDate    time.Time
	Venue        string
	Currency     string
	Lines        []InvoiceLine
	Subtotal     int64
	Discount     int64
	PromoCode    string
	Total        int64
	VATPercent   int
	VAT          int64
	Payment      PaymentMethod
	Status       string
	RefundAmount int64
}

func InvoiceNumber(o Order) string {
	year := o.CreatedAt.Year()
	if o.PaidAt != nil {
		year = o.PaidAt.Year()
	}
	return fmt.Sprintf("INV-%d-%08d", year, o.ID)
}

func VATIncluded(gross int64, percent int) int64 {
	if percent <= 0 {
		return 0
	}
	p := int64(percent)
	return (gross*p + (100+p)/2) / (100 + p)
}
