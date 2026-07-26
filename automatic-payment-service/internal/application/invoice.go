package application

import (
	"context"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type invoiceService struct {
	inv port.InvoiceRepository
}

func NewInvoiceService(inv port.InvoiceRepository) port.InvoiceUsecase {
	return &invoiceService{inv: inv}
}

func (s *invoiceService) GetInvoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	invoice, err := s.inv.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s *invoiceService) ListInvoicesByUser(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error) {
	invoices, err := s.inv.ListByUser(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Error("ListInvoicesByUser", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return invoices, nil
}

func (s *invoiceService) ListInvoicesBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Invoice, error) {
	invoices, err := s.inv.ListBySubscription(ctx, subscriptionID)
	if err != nil {
		logger.WithContext(ctx).Error("ListInvoicesBySubscription", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return invoices, nil
}
