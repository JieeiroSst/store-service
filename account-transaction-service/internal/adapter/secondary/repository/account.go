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

type accountRepo struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) port.AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(ctx context.Context, account *model.Account) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		logger.WithContext(ctx).Error("accountRepo.Create", zap.Error(err))
		return common.ErrDBFailed
	}
	return nil
}

func (r *accountRepo) GetByID(ctx context.Context, id string) (*model.Account, error) {
	var account model.Account
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&account)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		logger.WithContext(ctx).Error("accountRepo.GetByID", zap.Error(result.Error))
		return nil, common.ErrDBFailed
	}
	return &account, nil
}

func (r *accountRepo) List(ctx context.Context, p pagination.Page) ([]model.Account, error) {
	var accounts []model.Account
	if err := r.db.WithContext(ctx).
		Order("date_created desc").
		Offset(p.Offset).
		Limit(p.Limit).
		Find(&accounts).Error; err != nil {
		logger.WithContext(ctx).Error("accountRepo.List", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return accounts, nil
}
