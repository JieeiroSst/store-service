package application

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
)

type invoiceService struct {
	repo port.InvoiceRepository
}

func NewInvoiceService(repo port.InvoiceRepository) port.InvoiceUsecase {
	return &invoiceService{repo: repo}
}

func (s *invoiceService) Create(ctx context.Context, invoice *model.Invoice) error {
	return s.repo.Create(ctx, invoice)
}

func (s *invoiceService) Get(ctx context.Context, id int) (*model.Invoice, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *invoiceService) Update(ctx context.Context, invoice *model.Invoice) error {
	return s.repo.Update(ctx, invoice)
}

func (s *invoiceService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *invoiceService) List(ctx context.Context) ([]model.Invoice, error) {
	return s.repo.List(ctx)
}
