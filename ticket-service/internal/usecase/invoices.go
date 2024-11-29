package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/JIeeiroSst/ticket-service/internal/repository"
	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
	"golang.org/x/sync/errgroup"
)

type Invoices interface {
	ExportInvoices(ctx context.Context, customerID, ticketID int) error
	ExportPDFInvoices(ctx context.Context, invoiceDetails model.InvoiceDetails) error
}

type InvoicesUsecase struct {
	Repo            *repository.Repository
	UnidocSerectKey string
}

func NewInvoicesUsecase(Repo *repository.Repository,
	UnidocSerectKey string) *InvoicesUsecase {
	return &InvoicesUsecase{
		Repo:            Repo,
		UnidocSerectKey: UnidocSerectKey,
	}
}

func (u *InvoicesUsecase) ExportInvoices(ctx context.Context, customerID, ticketID int) error {
	var (
		customerCtx       *model.Customers
		invoiceDetailsCtx *model.InvoiceDetails
	)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		customer, err := u.Repo.Customers.Find(ctx, customerID)
		if err != nil {
			return err
		}
		customerCtx = customer
		return nil
	})

	g.Go(func() error {
		invoiceDetails, err := u.Repo.Invoices.FindInvoiceDetails(ctx, customerID, ticketID)
		if err != nil {
			return err
		}
		invoiceDetailsCtx = invoiceDetails
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if err := u.ExportPDFInvoices(ctx, invoiceDetailsCtx, customerCtx); err != nil {
		return err
	}
	return nil
}

func (u *InvoicesUsecase) ExportPDFInvoices(ctx context.Context, invoiceDetails *model.InvoiceDetails, customer *model.Customers) error {
	if invoiceDetails == nil || customer == nil {
		return errors.New("empty pointer")
	}
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

	err = c.WriteToFile(fmt.Sprintf("./docs/%v_%v.pdf", invoiceDetails.Tickets.TicketName, customer.CustomerName))
	if err != nil {
		logger.Error(err)
	}
	return nil
}
