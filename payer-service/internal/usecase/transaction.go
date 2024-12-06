package usecase

import (
	"context"

	"github.com/JIeeiroSst/payer-service/dto"
	"github.com/JIeeiroSst/payer-service/internal/repository"
	"github.com/JIeeiroSst/utils/logger"
)

type Transactions interface {
	Transactions(ctx context.Context, req dto.CreateTransactionsRequest) error
}

type TransactionUsecase struct {
	Repo *repository.Repository
}

func NewTransactionUsecase(repo *repository.Repository) *TransactionUsecase {
	return &TransactionUsecase{
		Repo: repo,
	}
}

func (u *TransactionUsecase) Transactions(ctx context.Context, req dto.CreateTransactionsRequest) error {
	buyer := req.Buyers.Build()
	payer := req.Payers.Build()
	transaction := req.Build()

	if err := u.Repo.Transactions.Transactions(ctx, payer, buyer, transaction.Amount,
		req.TransactionType, req.TransactionID, req.Description); err != nil {
		logger.Error(ctx, "error %v", err)
		return err
	}
	return nil
}
