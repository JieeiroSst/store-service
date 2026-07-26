package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/billing-service/common"
	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) port.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id int) (*model.Transaction, error) {
	var transaction model.Transaction
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", id).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	if err := r.db.WithContext(ctx).Model(&model.Transaction{}).
		Where("transaction_id = ?", transaction.TransactionID).
		Updates(transaction).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", id).Delete(&model.Transaction{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *transactionRepository) List(ctx context.Context) ([]model.Transaction, error) {
	var transactions []model.Transaction
	if err := r.db.WithContext(ctx).Find(&transactions).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return transactions, nil
}
