package repository

import (
	"context"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transactionRepo struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) port.TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *transactionRepo) ListBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]model.Transaction, error) {
	var txs []model.Transaction
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Find(&txs).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return txs, nil
}
