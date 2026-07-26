package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/billing-service/common"
	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
	"gorm.io/gorm"
)

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) port.InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *model.Invoice) error {
	if err := r.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *invoiceRepository) GetByID(ctx context.Context, id int) (*model.Invoice, error) {
	var invoice model.Invoice
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", id).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &invoice, nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *model.Invoice) error {
	if err := r.db.WithContext(ctx).Model(&model.Invoice{}).
		Where("invoice_id = ?", invoice.InvoiceID).
		Updates(invoice).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *invoiceRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", id).Delete(&model.Invoice{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *invoiceRepository) List(ctx context.Context) ([]model.Invoice, error) {
	var invoices []model.Invoice
	if err := r.db.WithContext(ctx).Find(&invoices).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return invoices, nil
}
