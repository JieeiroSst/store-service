package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type transactionService struct {
	repo port.TransactionRepository
}

func NewTransactionService(repo port.TransactionRepository) port.TransactionUsecase {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	if err := s.repo.Create(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *transactionService) GetTransaction(ctx context.Context, id int) (*model.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *transactionService) ListTransactions(ctx context.Context) ([]model.Transaction, error) {
	return s.repo.List(ctx)
}

func (s *transactionService) UpdateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	if _, err := s.repo.GetByID(ctx, transaction.TransactionID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *transactionService) DeleteTransaction(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
