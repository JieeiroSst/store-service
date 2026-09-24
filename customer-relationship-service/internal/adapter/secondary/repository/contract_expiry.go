package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"gorm.io/gorm"
)

type contractExpirer struct{ db *gorm.DB }

func NewContractExpirer(db *gorm.DB) port.ContractExpirer { return &contractExpirer{db: db} }

func (r *contractExpirer) ExpireOne(ctx context.Context, id uint, now time.Time) (bool, error) {
	res := conn(ctx, r.db).Model(&model.Contract{}).
		Where("id = ? AND end_date IS NOT NULL AND end_date <= ?", id, now).
		Where("contract_status IN ?", []string{model.ContractDraft, model.ContractPending, model.ContractApproved}).
		Update("contract_status", model.ContractExpired)
	if res.Error != nil {
		return false, common.ErrDBFailed
	}
	return res.RowsAffected > 0, nil
}

func (r *contractExpirer) ExpireDue(ctx context.Context, now time.Time) (int64, error) {
	res := conn(ctx, r.db).Model(&model.Contract{}).
		Where("end_date IS NOT NULL AND end_date < ?", now).
		Where("contract_status IN ?", []string{model.ContractDraft, model.ContractPending, model.ContractApproved}).
		Update("contract_status", model.ContractExpired)
	if res.Error != nil {
		return 0, common.ErrDBFailed
	}
	return res.RowsAffected, nil
}
