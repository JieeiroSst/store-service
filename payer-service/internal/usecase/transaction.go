package usecase

import (
	"context"

	"github.com/JIeeiroSst/payer-service/dto"
	"github.com/JIeeiroSst/payer-service/internal/repository"
	"github.com/JIeeiroSst/utils/logger"
)

type Transactions interface {
	Transactions(ctx context.Context, req dto.CreateTransactionsRequest) error
	GetTransactions(ctx context.Context, transactionID int) (*dto.Buyers, *dto.Payers, *dto.Transactions, error)
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

func (u *TransactionUsecase) GetTransactions(ctx context.Context,
	transactionID int) (*dto.Buyers, *dto.Payers, *dto.Transactions, error) {
	buyers, payers, transaction, err := u.Repo.Transactions.GetTransactions(ctx, transactionID)
	if err != nil {
		logger.Error(ctx, "error %v", err)
		return nil, nil, nil, err
	}
	transactionDTO, payersDTO, buyersDTo := dto.BuildGetTransaction(transaction, payers, buyers)
	return transactionDTO, payersDTO, buyersDTo, nil
}
