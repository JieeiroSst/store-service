package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/Jieeirosst/account-transaction-service/common"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/model"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/port"
	"github.com/Jieeirosst/account-transaction-service/pkg/pagination"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type transactionService struct {
	repo        port.TransactionRepository
	accountRepo port.AccountRepository
}

func NewTransactionService(repo port.TransactionRepository, accountRepo port.AccountRepository) port.TransactionUsecase {
	return &transactionService{repo: repo, accountRepo: accountRepo}
}

func (s *transactionService) create(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	lg := logger.WithContext(ctx)

	if tx.Amount <= 0 {
		return nil, common.ErrInvalidAmount
	}

	tx.ID = uuid.NewString()
	tx.DateCreated = time.Now()

	if err := s.repo.Create(ctx, tx); err != nil {
		lg.Error("create transaction", zap.Error(err), zap.String("type", string(tx.Type)))
		return nil, err
	}
	return tx, nil
}

func (s *transactionService) requireAccount(ctx context.Context, accountID string) error {
	_, err := s.accountRepo.GetByID(ctx, accountID)
	return err
}

func (s *transactionService) CreateDeposit(ctx context.Context, accountID string, amount float64) (*model.Transaction, error) {
	if err := s.requireAccount(ctx, accountID); err != nil {
		return nil, err
	}
	return s.create(ctx, &model.Transaction{
		Type:      model.TransactionTypeDeposit,
		Amount:    amount,
		AccountID: accountID,
	})
}

func (s *transactionService) CreateWithdrawal(ctx context.Context, accountID string, amount float64) (*model.Transaction, error) {
	if err := s.requireAccount(ctx, accountID); err != nil {
		return nil, err
	}
	return s.create(ctx, &model.Transaction{
		Type:      model.TransactionTypeWithdrawal,
		Amount:    amount,
		AccountID: accountID,
	})
}

func (s *transactionService) CreateTransfer(ctx context.Context, senderID, receiverID string, amount float64) (*model.Transaction, error) {
	if senderID == receiverID {
		return nil, common.ErrSameAccount
	}
	if err := s.requireAccount(ctx, senderID); err != nil {
		return nil, err
	}
	if err := s.requireAccount(ctx, receiverID); err != nil {
		return nil, err
	}
	return s.create(ctx, &model.Transaction{
		Type:       model.TransactionTypeTransfer,
		Amount:     amount,
		SenderID:   senderID,
		ReceiverID: receiverID,
	})
}

func (s *transactionService) CreatePayment(ctx context.Context, accountID, serviceName string, amount float64) (*model.Transaction, error) {
	if err := s.requireAccount(ctx, accountID); err != nil {
		return nil, err
	}
	return s.create(ctx, &model.Transaction{
		Type:        model.TransactionTypePayment,
		Amount:      amount,
		AccountID:   accountID,
		ServiceName: serviceName,
	})
}

func (s *transactionService) GetTransaction(ctx context.Context, id string) (*model.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *transactionService) ListTransactions(ctx context.Context, pageSize int32, pageToken string, txType string) ([]model.Transaction, string, error) {
	p := pagination.Parse(pageSize, pageToken)

	txs, err := s.repo.List(ctx, p, model.TransactionType(txType))
	if err != nil {
		return nil, "", err
	}
	return txs, p.NextToken(len(txs)), nil
}

func (s *transactionService) GetAccountTransactions(ctx context.Context, accountID string, pageSize int32, pageToken string) ([]model.Transaction, string, error) {
	p := pagination.Parse(pageSize, pageToken)

	txs, err := s.repo.ListByAccount(ctx, accountID, p)
	if err != nil {
		return nil, "", err
	}
	return txs, p.NextToken(len(txs)), nil
}
