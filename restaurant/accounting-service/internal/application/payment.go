package application

import (
	"context"

	"github.com/JIeeiroSst/accounting-service/internal/domain/model"
	"github.com/JIeeiroSst/accounting-service/internal/domain/port"
)

type paymentService struct {
	repo port.PaymentRepository
}

func NewPaymentService(repo port.PaymentRepository) port.PaymentUsecase {
	return &paymentService{repo: repo}
}

func (s *paymentService) MarkPaid(ctx context.Context, orderID int) error {
	return s.repo.UpdateStatus(ctx, orderID, model.PaymentStatusPaid)
}

func (s *paymentService) Get(ctx context.Context, orderID int) (*model.Payment, error) {
	return s.repo.FindByOrderID(ctx, orderID)
}
