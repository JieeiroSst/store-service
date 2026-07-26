package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type loanService struct {
	repo port.LoanRepository
}

func NewLoanService(repo port.LoanRepository) port.LoanUsecase {
	return &loanService{repo: repo}
}

func (s *loanService) CreateLoan(ctx context.Context, loan *model.Loan) (*model.Loan, error) {
	if err := s.repo.Create(ctx, loan); err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *loanService) GetLoan(ctx context.Context, id int) (*model.Loan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *loanService) ListLoans(ctx context.Context) ([]model.Loan, error) {
	return s.repo.List(ctx)
}

func (s *loanService) UpdateLoan(ctx context.Context, loan *model.Loan) (*model.Loan, error) {
	if _, err := s.repo.GetByID(ctx, loan.LoanID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, loan); err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *loanService) DeleteLoan(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
