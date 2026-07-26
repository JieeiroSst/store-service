package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) port.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, account *model.Account) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *accountRepository) GetByID(ctx context.Context, id int) (*model.Account, error) {
	var account model.Account
	err := r.db.WithContext(ctx).First(&account, "account_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &account, nil
}

func (r *accountRepository) List(ctx context.Context) ([]model.Account, error) {
	var accounts []model.Account
	if err := r.db.WithContext(ctx).Find(&accounts).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return accounts, nil
}

func (r *accountRepository) Update(ctx context.Context, account *model.Account) error {
	if err := r.db.WithContext(ctx).Model(&model.Account{}).Where("account_id = ?", account.AccountID).Updates(account).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *accountRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Account{}, "account_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
