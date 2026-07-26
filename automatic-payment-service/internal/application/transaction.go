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

type transactionService struct {
	txs port.TransactionRepository
}

func NewTransactionService(txs port.TransactionRepository) port.TransactionUsecase {
	return &transactionService{txs: txs}
}

func (s *transactionService) ListTransactionsBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Transaction, error) {
	txs, err := s.txs.ListBySubscription(ctx, subscriptionID)
	if err != nil {
		logger.WithContext(ctx).Error("ListTransactionsBySubscription", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return txs, nil
}
