package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type loanPaymentService struct {
	repo port.LoanPaymentRepository
}

func NewLoanPaymentService(repo port.LoanPaymentRepository) port.LoanPaymentUsecase {
	return &loanPaymentService{repo: repo}
}

func (s *loanPaymentService) CreateLoanPayment(ctx context.Context, payment *model.LoanPayment) (*model.LoanPayment, error) {
	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *loanPaymentService) GetLoanPayment(ctx context.Context, id int) (*model.LoanPayment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *loanPaymentService) ListLoanPayments(ctx context.Context) ([]model.LoanPayment, error) {
	return s.repo.List(ctx)
}

func (s *loanPaymentService) UpdateLoanPayment(ctx context.Context, payment *model.LoanPayment) (*model.LoanPayment, error) {
	if _, err := s.repo.GetByID(ctx, payment.LoanPaymentID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *loanPaymentService) DeleteLoanPayment(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
