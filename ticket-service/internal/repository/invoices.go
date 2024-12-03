package repository

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/pagination"
	"gorm.io/gorm"
)

type Invoices interface {
	SaveInvoices(ctx context.Context, invoices model.Invoices, invoiceDetails model.InvoiceDetails) error
	FindInvoiceDetails(ctx context.Context, customerID, ticketID int) (*model.InvoiceDetails, error)
	UpdateInvoiceDetails(ctx context.Context, status, ticketID int) error
	FindInvoices(ctx context.Context, p pagination.Pagination) (*pagination.Pagination, error)
}

type InvoicesRepository struct {
	db *gorm.DB
}

func NewInvoicesRepository(db *gorm.DB) *InvoicesRepository {
	return &InvoicesRepository{
		db: db,
	}
}

func (r *InvoicesRepository) SaveInvoices(ctx context.Context, invoices model.Invoices, invoiceDetails model.InvoiceDetails) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&invoices).Error; err != nil {
			logger.Error(ctx, "error %v", err)
			tx.Rollback()
			return err
		}
		if err := tx.Create(&invoiceDetails).Error; err != nil {
			logger.Error(ctx, "error %v", err)
			tx.Rollback()
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error(ctx, "error %v", err)
		return err
	}
	return nil
}

func (r *InvoicesRepository) FindInvoiceDetails(ctx context.Context, customerID, ticketID int) (*model.InvoiceDetails, error) {
	var invoiceDetails model.InvoiceDetails
	err := r.db.Where("ticket_id = ?", ticketID).
		Preload("Tickets").
		Preload("Invoices", "customer_id = ?", customerID).
		Find(&invoiceDetails).Error
	if err != nil {
		logger.Error(ctx, "error %v", err)
		return nil, err
	}
	return &invoiceDetails, nil
}

func (r *InvoicesRepository) UpdateInvoiceDetails(ctx context.Context, status, ticketID int) error {
	if err := r.db.Model(model.InvoiceDetails{}).Where("ticket_id = ?", ticketID).
		Update("status = ?", status).Error; err != nil {
		logger.Error(ctx, "error %v", err)
		return err
	}
	return nil
}

func (r *InvoicesRepository) FindInvoices(ctx context.Context, param pagination.Pagination) (*pagination.Pagination, error) {
	var invoices []model.Invoices

	r.db.Scopes(pagination.Paginate(invoices, &param, r.db)).Find(&invoices)
	param.Rows = invoices

	return &param, nil
}
