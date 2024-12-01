package build

import (
	"time"

	"github.com/JIeeiroSst/ticket-service/common"
	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
)

func BuildInvoiceDetails(customerCtx *model.Customers,
	ticketCtx *model.Tickets, quantity int) (model.Invoices, model.InvoiceDetails) {
	var (
		invoices       model.Invoices
		invoiceDetails model.InvoiceDetails
	)

	if customerCtx == nil || ticketCtx == nil {
		return invoices, invoiceDetails
	}

	invoicesID := geared_id.GearedIntID()
	invoices = model.Invoices{
		InvoicesID:  invoicesID,
		CustomerID:  customerCtx.CustomerID,
		BuyDate:     time.Now(),
		TotalAmount: float64(quantity) * ticketCtx.Amount,
		Note:        "",
	}

	invoiceDetails = model.InvoiceDetails{
		InvoiceDetailID: geared_id.GearedIntID(),
		InvoicesID:      invoicesID,
		TicketID:        ticketCtx.TicketID,
		Quantity:        quantity,
		Status:          common.PENDING.Value(),
	}

	return invoices, invoiceDetails
}
