package service

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type InvoiceOptions struct {
	Seller     domain.Seller
	VATPercent int
}

func (s *OrderService) Invoice(ctx context.Context, actor inbound.Principal, id int64) (domain.Invoice, error) {
	x, _, err := s.load(ctx, actor, id)
	if err != nil {
		return domain.Invoice{}, err
	}
	if x.PaidAt == nil || (x.Status != domain.OrderPaid && x.Status != domain.OrderRefunded) {
		return domain.Invoice{}, fmt.Errorf("%w: only a paid order has an invoice", domain.ErrConflict)
	}
	e, err := s.events.Get(ctx, x.EventID)
	if err != nil {
		return domain.Invoice{}, err
	}
	inv := domain.Invoice{
		Number: domain.InvoiceNumber(x), IssuedAt: *x.PaidAt, OrderID: x.ID, Seller: s.opts.Invoice.Seller,
		BuyerName: x.BuyerName, BuyerEmail: x.BuyerEmail, EventTitle: titleOf(e, showtime(e, x.SessionID)), EventDate: showtime(e, x.SessionID).StartsAt, Venue: e.Venue,
		Currency: x.Currency, Subtotal: x.Subtotal, Discount: x.Discount, PromoCode: x.PromoCode, Total: x.Total,
		VATPercent: s.opts.Invoice.VATPercent, VAT: domain.VATIncluded(x.Total, s.opts.Invoice.VATPercent),
		Payment: x.PaymentMethod, Status: "paid", RefundAmount: x.RefundAmount,
	}
	if x.Status == domain.OrderRefunded {
		inv.Status = "refunded"
	}
	for _, it := range x.Items {
		inv.Lines = append(inv.Lines, domain.InvoiceLine{Description: it.Name, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: it.UnitPrice * int64(it.Quantity)})
	}
	return inv, nil
}
