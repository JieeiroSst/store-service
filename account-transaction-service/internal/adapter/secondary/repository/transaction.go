package repository

import (
	"context"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/Jieeirosst/account-transaction-service/common"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/model"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/port"
	"github.com/Jieeirosst/account-transaction-service/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type transactionRepo struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) port.TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(ctx context.Context, tx *model.Transaction) error {
	if err := r.db.WithContext(ctx).Create(tx).Error; err != nil {
		logger.WithContext(ctx).Error("transactionRepo.Create", zap.Error(err))
		return common.ErrDBFailed
	}
	return nil
}

func (r *transactionRepo) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	var tx model.Transaction
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&tx)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		logger.WithContext(ctx).Error("transactionRepo.GetByID", zap.Error(result.Error))
		return nil, common.ErrDBFailed
	}
	return &tx, nil
}

func (r *transactionRepo) List(ctx context.Context, p pagination.Page, txType model.TransactionType) ([]model.Transaction, error) {
	q := r.db.WithContext(ctx).Order("date_created desc")
	if txType != "" {
		q = q.Where("type = ?", txType)
	}

	var txs []model.Transaction
	if err := q.Offset(p.Offset).Limit(p.Limit).Find(&txs).Error; err != nil {
		logger.WithContext(ctx).Error("transactionRepo.List", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return txs, nil
}

func (r *transactionRepo) ListByAccount(ctx context.Context, accountID string, p pagination.Page) ([]model.Transaction, error) {
	var txs []model.Transaction
	if err := r.db.WithContext(ctx).
		Where("account_id = ? OR sender_id = ? OR receiver_id = ?", accountID, accountID, accountID).
		Order("date_created desc").
		Offset(p.Offset).
		Limit(p.Limit).
		Find(&txs).Error; err != nil {
		logger.WithContext(ctx).Error("transactionRepo.ListByAccount", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return txs, nil
}
