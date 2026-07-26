package repository

import (
	"context"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type invoiceRepo struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) port.InvoiceRepository {
	return &invoiceRepo{db: db}
}

func (r *invoiceRepo) Create(ctx context.Context, inv *model.Invoice) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *invoiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	var inv model.Invoice
	err := r.db.WithContext(ctx).Where("invoice_id = ?", id).First(&inv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &inv, nil
}

func (r *invoiceRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error) {
	var invoices []model.Invoice
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return invoices, nil
}

func (r *invoiceRepo) ListBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Invoice, error) {
	var invoices []model.Invoice
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return invoices, nil
}
