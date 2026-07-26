package repository

import (
	"context"
	"errors"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

const mysqlDuplicateErrNo = 1062

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) port.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *model.Transaction) error {
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		var mysqlErr *mysqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateErrNo {
			return common.ErrDuplicate
		}
		return common.ErrDBFailed
	}
	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id int) (*model.Transaction, error) {
	var transaction model.Transaction
	err := r.db.WithContext(ctx).First(&transaction, "transaction_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &transaction, nil
}

func (r *transactionRepository) List(ctx context.Context) ([]model.Transaction, error) {
	var transactions []model.Transaction
	if err := r.db.WithContext(ctx).Find(&transactions).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return transactions, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *model.Transaction) error {
	if err := r.db.WithContext(ctx).Model(&model.Transaction{}).Where("transaction_id = ?", transaction.TransactionID).Updates(transaction).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Transaction{}, "transaction_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
