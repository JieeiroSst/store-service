package mfa

import (
	"context"
	"errors"
	"time"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context, userID string) (model.UserTOTP, error) {
	var rec model.UserTOTP
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.UserTOTP{}, common.ErrMFANotSetUp
	}
	return rec, err
}

func (r *Repository) SavePending(ctx context.Context, userID, sealedSecret string) error {
	now := time.Now()
	rec := model.UserTOTP{UserID: userID, Secret: sealedSecret, CreatedAt: now, UpdatedAt: now}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{"secret": sealedSecret, "last_used_step": 0, "recovery_codes": "", "updated_at": now}),
		Where:     clause.Where{Exprs: []clause.Expression{clause.Eq{Column: clause.Column{Table: "user_totps", Name: "enabled"}, Value: false}}},
	}).Create(&rec)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrMFAAlreadyEnabled
	}
	return nil
}

func (r *Repository) Enable(ctx context.Context, userID string, step int64, recoveryCodes string, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.UserTOTP{}).
		Where("user_id = ? AND enabled = ? AND last_used_step < ?", userID, false, step).
		Updates(map[string]any{"enabled": true, "enabled_at": at, "last_used_step": step, "recovery_codes": recoveryCodes, "updated_at": at})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrMFAInvalidCode
	}
	return nil
}

func (r *Repository) MarkUsed(ctx context.Context, userID string, step int64) (bool, error) {
	res := r.db.WithContext(ctx).Model(&model.UserTOTP{}).
		Where("user_id = ? AND last_used_step < ?", userID, step).
		Updates(map[string]any{"last_used_step": step, "updated_at": time.Now()})
	return res.RowsAffected == 1, res.Error
}

func (r *Repository) SwapRecoveryCodes(ctx context.Context, userID, old, new string) (bool, error) {
	res := r.db.WithContext(ctx).Model(&model.UserTOTP{}).
		Where("user_id = ? AND enabled = ? AND recovery_codes = ?", userID, true, old).
		Updates(map[string]any{"recovery_codes": new, "updated_at": time.Now()})
	return res.RowsAffected == 1, res.Error
}

func (r *Repository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&model.UserTOTP{}, "user_id = ?", userID).Error
}
