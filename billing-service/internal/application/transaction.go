package application

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
)

type transactionService struct {
	repo port.TransactionRepository
}

func NewTransactionService(repo port.TransactionRepository) port.TransactionUsecase {
	return &transactionService{repo: repo}
}

func (s *transactionService) Create(ctx context.Context, transaction *model.Transaction) error {
	return s.repo.Create(ctx, transaction)
}

func (s *transactionService) Get(ctx context.Context, id int) (*model.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *transactionService) Update(ctx context.Context, transaction *model.Transaction) error {
	return s.repo.Update(ctx, transaction)
}

func (s *transactionService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *transactionService) List(ctx context.Context) ([]model.Transaction, error) {
	return s.repo.List(ctx)
}
