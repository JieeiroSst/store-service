package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/JIeeiroSst/ticket-service/internal/repository"
	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
)

type Invoices interface {
	ExportInvoices(ctx context.Context, customerID, ticketID int) error
	ExportPDFInvoices(ctx context.Context, invoiceDetails model.InvoiceDetails) error
}

type InvoicesUsecase struct {
	InvoicesRepo    repository.Invoices
	UnidocSerectKey string
}

func NewInvoicesUsecase(InvoicesRepo repository.Invoices,
	UnidocSerectKey string) *InvoicesUsecase {
	return &InvoicesUsecase{
		InvoicesRepo:    InvoicesRepo,
		UnidocSerectKey: UnidocSerectKey,
	}
}

func (u *InvoicesUsecase) ExportInvoices(ctx context.Context, customerID, ticketID int) error {

	return nil
}

func (u *InvoicesUsecase) ExportPDFInvoices(ctx context.Context, invoiceDetails model.InvoiceDetails, customer model.Customers) error {
	err := license.SetMeteredKey(u.UnidocSerectKey)
	if err != nil {
		return err
	}

	c := creator.New()
	c.NewPage()
	invoice := c.NewInvoice()

	invoice.SetSellerAddress(&creator.InvoiceAddress{
		Name:   customer.CustomerName,
		Street: customer.Address,
		Phone:  customer.Phone,
		Email:  customer.Email,
	})

	invoice.AddLine(
		invoiceDetails.Tickets.TicketName,
		fmt.Sprintf("%v", invoiceDetails.Tickets.Quantity),
		fmt.Sprintf("%v", invoiceDetails.Tickets.Amount),
		fmt.Sprintf("%v", invoiceDetails.Tickets.Amount),
	)

	invoice.SetTotal(fmt.Sprintf("$%v", invoiceDetails.Invoices.TotalAmount))

	if err := c.Draw(invoice); err != nil {
		log.Fatalf("Error drawing: %v", err)
	}

	err = c.WriteToFile(fmt.Sprintf("%v_%v.pdf", invoiceDetails.Tickets.TicketName, customer.CustomerName))
	if err != nil {
		logger.Error(err)
	}
	return nil
}
