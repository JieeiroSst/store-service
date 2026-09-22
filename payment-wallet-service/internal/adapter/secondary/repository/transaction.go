package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) port.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) GetByID(ctx context.Context, transactionID string) (*model.Transaction, error) {
	var txn model.Transaction
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&txn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &txn, nil
}

func (r *transactionRepository) GetByReferenceID(ctx context.Context, walletID, referenceID string) (*model.Transaction, error) {
	if referenceID == "" {
		return nil, port.ErrNotFound
	}
	var txn model.Transaction
	err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND reference_id = ?", walletID, referenceID).
		First(&txn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &txn, nil
}

func (r *transactionRepository) ListByWallet(ctx context.Context, walletID string, limit, offset int) ([]model.Transaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var txns []model.Transaction
	err := r.db.WithContext(ctx).
		Where("wallet_id = ?", walletID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&txns).Error
	return txns, err
}

func (r *transactionRepository) ListByWalletAndDateRange(ctx context.Context, walletID string, from, to time.Time) ([]model.Transaction, error) {
	var txns []model.Transaction
	err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND created_at >= ? AND created_at < ?", walletID, from, to).
		Order("created_at ASC").
		Find(&txns).Error
	return txns, err
}

func (r *transactionRepository) SumOutgoingSince(ctx context.Context, walletID string, since time.Time) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("wallet_id = ? AND status = ? AND type IN ? AND created_at >= ?",
			walletID, model.TxnCompleted, []model.TransactionType{model.TxnWithdraw, model.TxnTransferOut}, since).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
